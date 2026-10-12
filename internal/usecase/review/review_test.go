package review

import (
	"context"
	"errors"
	"testing"

	"labshop/internal/domain/catalog"
	"labshop/internal/domain/money"
	domainorder "labshop/internal/domain/order"
	domainreview "labshop/internal/domain/review"
	"labshop/internal/usecase/port"
)

type fakeReviewRepo struct {
	reviews map[string]*domainreview.Review
}

func newFakeReviewRepo() *fakeReviewRepo {
	return &fakeReviewRepo{reviews: make(map[string]*domainreview.Review)}
}

func (f *fakeReviewRepo) Create(ctx context.Context, r *domainreview.Review) error {
	f.reviews[r.ID] = r
	return nil
}

func (f *fakeReviewRepo) GetByID(ctx context.Context, id string) (*domainreview.Review, error) {
	r, ok := f.reviews[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return r, nil
}

func (f *fakeReviewRepo) GetByUserAndProduct(ctx context.Context, userID, productID string) (*domainreview.Review, error) {
	for _, r := range f.reviews {
		if r.UserID == userID && r.ProductID == productID {
			return r, nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeReviewRepo) ListByProduct(ctx context.Context, productID string, limit, offset int) ([]domainreview.Review, int, error) {
	var list []domainreview.Review
	for _, r := range f.reviews {
		if r.ProductID == productID {
			list = append(list, *r)
		}
	}
	return list, len(list), nil
}

func (f *fakeReviewRepo) Delete(ctx context.Context, id string) error {
	delete(f.reviews, id)
	return nil
}

func (f *fakeReviewRepo) GetProductRatingStats(ctx context.Context, productID string) (float64, int, error) {
	var (
		sum   int
		count int
	)
	for _, r := range f.reviews {
		if r.ProductID == productID {
			sum += r.Rating
			count++
		}
	}
	if count == 0 {
		return 0, 0, nil
	}
	return float64(sum) / float64(count), count, nil
}

type fakeOrderRepo struct {
	hasDelivered bool
}

func (f *fakeOrderRepo) Create(ctx context.Context, o *domainorder.Order) error { return nil }
func (f *fakeOrderRepo) GetByID(ctx context.Context, id string) (*domainorder.Order, error) {
	return nil, nil
}
func (f *fakeOrderRepo) GetByIdempotencyKey(ctx context.Context, userID, key string) (*domainorder.Order, error) {
	return nil, nil
}
func (f *fakeOrderRepo) UpdateStatus(ctx context.Context, id string, status domainorder.Status, paymentRef string) error {
	return nil
}
func (f *fakeOrderRepo) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]domainorder.Order, int, error) {
	return nil, 0, nil
}
func (f *fakeOrderRepo) ListAll(ctx context.Context, limit, offset int) ([]domainorder.Order, int, error) {
	return nil, 0, nil
}
func (f *fakeOrderRepo) HasDeliveredOrderWithProduct(ctx context.Context, userID, productID string) (bool, error) {
	return f.hasDelivered, nil
}

type fakeCatalogRepo struct {
	product *catalog.Product
}

func (f *fakeCatalogRepo) ListCategories(ctx context.Context) ([]catalog.Category, error) {
	return nil, nil
}
func (f *fakeCatalogRepo) GetCategoryByID(ctx context.Context, id string) (*catalog.Category, error) {
	return nil, nil
}
func (f *fakeCatalogRepo) CreateCategory(ctx context.Context, c *catalog.Category) error {
	return nil
}
func (f *fakeCatalogRepo) ListProducts(ctx context.Context, filter port.ProductFilter) ([]catalog.Product, int, error) {
	return nil, 0, nil
}
func (f *fakeCatalogRepo) GetProductByID(ctx context.Context, id string) (*catalog.Product, error) {
	return f.product, nil
}
func (f *fakeCatalogRepo) GetProductBySlug(ctx context.Context, slug string) (*catalog.Product, error) {
	if f.product != nil && f.product.Slug == slug {
		return f.product, nil
	}
	return nil, errors.New("not found")
}
func (f *fakeCatalogRepo) CreateProduct(ctx context.Context, p *catalog.Product) error { return nil }
func (f *fakeCatalogRepo) UpdateProduct(ctx context.Context, p *catalog.Product) error { return nil }
func (f *fakeCatalogRepo) DeductStockIfAvailable(ctx context.Context, id string, quantity int) (bool, error) {
	return true, nil
}
func (f *fakeCatalogRepo) RestoreStock(ctx context.Context, id string, quantity int) error { return nil }
func (f *fakeCatalogRepo) UpdateProductRating(ctx context.Context, id string, ratingAvg float64, ratingCount int) error {
	if f.product != nil {
		f.product.RatingAvg = ratingAvg
		f.product.RatingCount = ratingCount
	}
	return nil
}

type fakeIDGen struct{}

func (f fakeIDGen) NewID() string { return "rev-123" }

func TestReviewUsecase(t *testing.T) {
	reviewRepo := newFakeReviewRepo()
	orderRepo := &fakeOrderRepo{hasDelivered: false}
	catRepo := &fakeCatalogRepo{
		product: &catalog.Product{
			ID:    "p-1",
			Name:  "Mechanical Keyboard",
			Slug:  "mechanical-keyboard",
			Price: money.VND(1000000),
		},
	}
	idGen := fakeIDGen{}

	uc := New(reviewRepo, orderRepo, catRepo, idGen)
	ctx := context.Background()

	input := CreateReviewInput{
		ProductSlug: "mechanical-keyboard",
		Rating:      5,
		Comment:     "Awesome typing feel!",
	}

	// Without delivered order -> fails
	_, err := uc.CreateReview(ctx, "u-1", "Alice", input)
	if err == nil {
		t.Fatalf("expected error when user has no delivered order")
	}

	// With delivered order -> succeeds
	orderRepo.hasDelivered = true
	rev, err := uc.CreateReview(ctx, "u-1", "Alice", input)
	if err != nil {
		t.Fatalf("unexpected error creating review: %v", err)
	}
	if rev.Rating != 5 || catRepo.product.RatingAvg != 5.0 || catRepo.product.RatingCount != 1 {
		t.Fatalf("review creation did not update rating correctly: %+v, %+v", rev, catRepo.product)
	}

	// Duplicate review -> conflict
	_, err = uc.CreateReview(ctx, "u-1", "Alice", input)
	if err == nil {
		t.Fatalf("expected conflict on duplicate review")
	}

	// List reviews
	list, total, err := uc.ListReviews(ctx, "mechanical-keyboard", 10, 0)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("failed to list reviews: %v, total: %d", err, total)
	}

	// Delete review by another user -> forbidden
	err = uc.DeleteReview(ctx, "u-other", rev.ID, false)
	if err == nil {
		t.Fatalf("expected error deleting another user's review")
	}

	// Delete review by owner -> success
	err = uc.DeleteReview(ctx, "u-1", rev.ID, false)
	if err != nil {
		t.Fatalf("failed to delete review: %v", err)
	}
	if catRepo.product.RatingCount != 0 {
		t.Fatalf("expected rating count 0 after delete")
	}

	// Delete non-existent review
	err = uc.DeleteReview(ctx, "u-1", "non-existent", false)
	if err == nil {
		t.Fatalf("expected error deleting non-existent review")
	}

	// Create review for non-existent product
	_, err = uc.CreateReview(ctx, "u-1", "Alice", CreateReviewInput{ProductSlug: "non-existent"})
	if err == nil {
		t.Fatalf("expected error creating review for non-existent product")
	}
}
