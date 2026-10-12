package review

import (
	"context"
	"strings"
	"time"

	"labshop/internal/domain/errs"
	domainreview "labshop/internal/domain/review"
	"labshop/internal/usecase/port"
)

type CreateReviewInput struct {
	ProductSlug string
	Rating      int
	Comment     string
}

type Usecase struct {
	reviewRepo  port.ReviewRepository
	orderRepo   port.OrderRepository
	catalogRepo port.CatalogRepository
	idGen       port.IDGenerator
}

func New(
	reviewRepo port.ReviewRepository,
	orderRepo port.OrderRepository,
	catalogRepo port.CatalogRepository,
	idGen port.IDGenerator,
) *Usecase {
	return &Usecase{
		reviewRepo:  reviewRepo,
		orderRepo:   orderRepo,
		catalogRepo: catalogRepo,
		idGen:       idGen,
	}
}

func (u *Usecase) ListReviews(ctx context.Context, productSlug string, limit, offset int) ([]domainreview.Review, int, error) {
	prod, err := u.catalogRepo.GetProductBySlug(ctx, productSlug)
	if err != nil || prod == nil {
		return nil, 0, errs.NewNotFound("product", productSlug)
	}

	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return u.reviewRepo.ListByProduct(ctx, prod.ID, limit, offset)
}

func (u *Usecase) CreateReview(ctx context.Context, userID, userName string, in CreateReviewInput) (*domainreview.Review, error) {
	prod, err := u.catalogRepo.GetProductBySlug(ctx, in.ProductSlug)
	if err != nil || prod == nil {
		return nil, errs.NewNotFound("product", in.ProductSlug)
	}

	hasDelivered, err := u.orderRepo.HasDeliveredOrderWithProduct(ctx, userID, prod.ID)
	if err != nil || !hasDelivered {
		return nil, errs.NewBusinessRule("no_delivered_order", "reviews can only be submitted for delivered purchases")
	}

	existing, _ := u.reviewRepo.GetByUserAndProduct(ctx, userID, prod.ID)
	if existing != nil {
		return nil, errs.NewConflict("you have already reviewed this product")
	}

	now := time.Now()
	rev := &domainreview.Review{
		ID:        u.idGen.NewID(),
		UserID:    userID,
		ProductID: prod.ID,
		UserName:  userName,
		Rating:    in.Rating,
		Comment:   strings.TrimSpace(in.Comment),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := rev.Validate(); err != nil {
		return nil, err
	}

	if err := u.reviewRepo.Create(ctx, rev); err != nil {
		return nil, err
	}

	// Recalculate stats
	avg, count, errStats := u.reviewRepo.GetProductRatingStats(ctx, prod.ID)
	if errStats == nil {
		_ = u.catalogRepo.UpdateProductRating(ctx, prod.ID, avg, count)
	}

	return rev, nil
}

func (u *Usecase) DeleteReview(ctx context.Context, userID, reviewID string, isAdmin bool) error {
	rev, err := u.reviewRepo.GetByID(ctx, reviewID)
	if err != nil || rev == nil {
		return errs.NewNotFound("review", reviewID)
	}

	if !isAdmin && rev.UserID != userID {
		return errs.NewForbidden("cannot delete another user's review")
	}

	if err := u.reviewRepo.Delete(ctx, reviewID); err != nil {
		return err
	}

	// Recalculate stats
	avg, count, errStats := u.reviewRepo.GetProductRatingStats(ctx, rev.ProductID)
	if errStats == nil {
		_ = u.catalogRepo.UpdateProductRating(ctx, rev.ProductID, avg, count)
	}

	return nil
}
