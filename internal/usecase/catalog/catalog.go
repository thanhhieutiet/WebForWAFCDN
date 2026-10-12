package catalog

import (
	"context"
	"strings"

	"labshop/internal/domain/catalog"
	"labshop/internal/domain/errs"
	"labshop/internal/usecase/port"
)

type Usecase struct {
	catalogRepo port.CatalogRepository
}

func New(repo port.CatalogRepository) *Usecase {
	return &Usecase{
		catalogRepo: repo,
	}
}

func (u *Usecase) ListCategories(ctx context.Context) ([]catalog.Category, error) {
	return u.catalogRepo.ListCategories(ctx)
}

func (u *Usecase) ListProducts(ctx context.Context, filter port.ProductFilter) ([]catalog.Product, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	} else if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	filter.Query = strings.TrimSpace(filter.Query)
	filter.CategoryID = strings.TrimSpace(filter.CategoryID)

	return u.catalogRepo.ListProducts(ctx, filter)
}

func (u *Usecase) GetProductBySlug(ctx context.Context, slug string) (*catalog.Product, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, errs.NewValidation("slug is required", errs.FieldViolation{Field: "slug", Issue: "cannot be empty"})
	}

	prod, err := u.catalogRepo.GetProductBySlug(ctx, slug)
	if err != nil || prod == nil {
		return nil, errs.NewNotFound("product", slug)
	}

	return prod, nil
}
