package cart

import (
	"context"
	"errors"
	"testing"
	"time"

	domaincart "labshop/internal/domain/cart"
	"labshop/internal/domain/catalog"
	"labshop/internal/domain/coupon"
	"labshop/internal/domain/money"
	"labshop/internal/usecase/port"
)

type fakeCartRepo struct {
	carts map[string]*domaincart.Cart
}

func newFakeCartRepo() *fakeCartRepo {
	return &fakeCartRepo{carts: make(map[string]*domaincart.Cart)}
}

func (f *fakeCartRepo) GetByUserID(ctx context.Context, userID string) (*domaincart.Cart, error) {
	c, ok := f.carts[userID]
	if !ok {
		return nil, errors.New("not found")
	}
	return c, nil
}

func (f *fakeCartRepo) Save(ctx context.Context, c *domaincart.Cart) error {
	f.carts[c.UserID] = c
	return nil
}

func (f *fakeCartRepo) DeleteByUserID(ctx context.Context, userID string) error {
	delete(f.carts, userID)
	return nil
}

type fakeCouponRepo struct {
	coupons map[string]*coupon.Coupon
}

func newFakeCouponRepo() *fakeCouponRepo {
	return &fakeCouponRepo{coupons: make(map[string]*coupon.Coupon)}
}

func (f *fakeCouponRepo) GetByCode(ctx context.Context, code string) (*coupon.Coupon, error) {
	c, ok := f.coupons[code]
	if !ok {
		return nil, errors.New("coupon not found")
	}
	return c, nil
}

func (f *fakeCouponRepo) IncrementUsage(ctx context.Context, code string) error {
	if c, ok := f.coupons[code]; ok {
		c.UsedCount++
	}
	return nil
}

func (f *fakeCouponRepo) List(ctx context.Context) ([]coupon.Coupon, error) {
	var list []coupon.Coupon
	for _, c := range f.coupons {
		list = append(list, *c)
	}
	return list, nil
}

func (f *fakeCouponRepo) Save(ctx context.Context, c *coupon.Coupon) error {
	f.coupons[c.Code] = c
	return nil
}

type fakeCatalogRepo struct {
	products map[string]*catalog.Product
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
	p, ok := f.products[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return p, nil
}
func (f *fakeCatalogRepo) GetProductBySlug(ctx context.Context, slug string) (*catalog.Product, error) {
	return nil, nil
}
func (f *fakeCatalogRepo) CreateProduct(ctx context.Context, p *catalog.Product) error {
	return nil
}
func (f *fakeCatalogRepo) UpdateProduct(ctx context.Context, p *catalog.Product) error {
	return nil
}
func (f *fakeCatalogRepo) DeductStockIfAvailable(ctx context.Context, id string, quantity int) (bool, error) {
	return true, nil
}
func (f *fakeCatalogRepo) RestoreStock(ctx context.Context, id string, quantity int) error {
	return nil
}
func (f *fakeCatalogRepo) UpdateProductRating(ctx context.Context, id string, ratingAvg float64, ratingCount int) error {
	return nil
}

func TestCartUsecase(t *testing.T) {
	cartRepo := newFakeCartRepo()
	couponRepo := newFakeCouponRepo()
	catalogRepo := &fakeCatalogRepo{products: make(map[string]*catalog.Product)}

	prod1 := &catalog.Product{
		ID:     "prod-1",
		Name:   "Keyboard",
		Price:  money.VND(100000),
		Stock:  5,
		Status: catalog.StatusActive,
	}
	catalogRepo.products[prod1.ID] = prod1

	coup1 := &coupon.Coupon{
		Code:        "DISCOUNT10",
		Type:        coupon.TypePercent,
		Value:       10,
		MinSubtotal: money.VND(50000),
		ValidFrom:   time.Now().Add(-time.Hour),
		ValidTo:     time.Now().Add(time.Hour),
	}
	couponRepo.coupons[coup1.Code] = coup1

	uc := New(cartRepo, catalogRepo, couponRepo)
	ctx := context.Background()

	// Update item quantity
	cv, err := uc.UpdateItemQuantity(ctx, "user-1", "prod-1", 2)
	if err != nil {
		t.Fatalf("unexpected error updating quantity: %v", err)
	}
	if cv.ItemCount != 2 || cv.Subtotal != money.VND(200000) {
		t.Fatalf("unexpected cart state: %+v", cv)
	}
	// Subtotal 200,000 < 500,000 => shipping 30,000 => total 230,000
	if cv.ShippingFee != money.VND(30000) || cv.Total != money.VND(230000) {
		t.Fatalf("unexpected total: %v, shipping: %v", cv.Total, cv.ShippingFee)
	}

	// Apply coupon
	cv, err = uc.ApplyCoupon(ctx, "user-1", "DISCOUNT10")
	if err != nil {
		t.Fatalf("failed to apply coupon: %v", err)
	}
	// 10% of 200,000 = 20,000 discount. Shipping 30,000. Total = 210,000
	if cv.Discount != money.VND(20000) || cv.Total != money.VND(210000) {
		t.Fatalf("unexpected discount %v or total %v", cv.Discount, cv.Total)
	}

	// Remove item
	cv, err = uc.RemoveItem(ctx, "user-1", "prod-1")
	if err != nil || len(cv.Items) != 0 {
		t.Fatalf("failed to remove item")
	}

	// Update quantity to 0 removes item
	_, _ = uc.UpdateItemQuantity(ctx, "user-1", "prod-1", 1)
	cv, err = uc.UpdateItemQuantity(ctx, "user-1", "prod-1", 0)
	if err != nil || len(cv.Items) != 0 {
		t.Fatalf("expected quantity 0 to remove item")
	}

	// Update non-existent product
	_, err = uc.UpdateItemQuantity(ctx, "user-1", "prod-non-existent", 1)
	if err == nil {
		t.Fatalf("expected error for non-existent product")
	}

	// Remove coupon
	_, _ = uc.ApplyCoupon(ctx, "user-1", "DISCOUNT10")
	cv, err = uc.RemoveCoupon(ctx, "user-1")
	if err != nil || cv.CouponCode != "" {
		t.Fatalf("failed to remove coupon")
	}

	// Apply empty coupon
	_, err = uc.ApplyCoupon(ctx, "user-1", "")
	if err == nil {
		t.Fatalf("expected error for empty coupon code")
	}

	// Apply non-existent coupon
	_, err = uc.ApplyCoupon(ctx, "user-1", "NONEXISTENT")
	if err == nil {
		t.Fatalf("expected error for non-existent coupon")
	}

	// Clear cart
	err = uc.ClearCart(ctx, "user-1")
	if err != nil {
		t.Fatalf("failed to clear cart: %v", err)
	}
}
