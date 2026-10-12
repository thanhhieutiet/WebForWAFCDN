package catalog

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	domaincatalog "labshop/internal/domain/catalog"
	"labshop/internal/domain/money"
	"labshop/internal/usecase/port"
)

type fakeCatalogRepo struct {
	categories []domaincatalog.Category
	products   map[string]*domaincatalog.Product
}

func newFakeCatalogRepo() *fakeCatalogRepo {
	return &fakeCatalogRepo{
		categories: make([]domaincatalog.Category, 0),
		products:   make(map[string]*domaincatalog.Product),
	}
}

func (f *fakeCatalogRepo) ListCategories(ctx context.Context) ([]domaincatalog.Category, error) {
	return f.categories, nil
}

func (f *fakeCatalogRepo) GetCategoryByID(ctx context.Context, id string) (*domaincatalog.Category, error) {
	for _, c := range f.categories {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, errors.New("category not found")
}

func (f *fakeCatalogRepo) CreateCategory(ctx context.Context, c *domaincatalog.Category) error {
	f.categories = append(f.categories, *c)
	return nil
}

func (f *fakeCatalogRepo) ListProducts(ctx context.Context, filter port.ProductFilter) ([]domaincatalog.Product, int, error) {
	var res []domaincatalog.Product
	for _, p := range f.products {
		if filter.CategoryID != "" && p.CategoryID != filter.CategoryID {
			continue
		}
		if filter.Query != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(filter.Query)) {
			continue
		}
		res = append(res, *p)
	}
	return res, len(res), nil
}

func (f *fakeCatalogRepo) GetProductByID(ctx context.Context, id string) (*domaincatalog.Product, error) {
	p, ok := f.products[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return p, nil
}

func (f *fakeCatalogRepo) GetProductBySlug(ctx context.Context, slug string) (*domaincatalog.Product, error) {
	for _, p := range f.products {
		if p.Slug == slug {
			return p, nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeCatalogRepo) CreateProduct(ctx context.Context, p *domaincatalog.Product) error {
	f.products[p.ID] = p
	return nil
}

func (f *fakeCatalogRepo) UpdateProduct(ctx context.Context, p *domaincatalog.Product) error {
	f.products[p.ID] = p
	return nil
}

func (f *fakeCatalogRepo) DeductStockIfAvailable(ctx context.Context, id string, quantity int) (bool, error) {
	p, ok := f.products[id]
	if !ok || p.Stock < quantity {
		return false, nil
	}
	p.Stock -= quantity
	return true, nil
}

func (f *fakeCatalogRepo) RestoreStock(ctx context.Context, id string, quantity int) error {
	if p, ok := f.products[id]; ok {
		p.Stock += quantity
	}
	return nil
}

func (f *fakeCatalogRepo) UpdateProductRating(ctx context.Context, id string, ratingAvg float64, ratingCount int) error {
	if p, ok := f.products[id]; ok {
		p.RatingAvg = ratingAvg
		p.RatingCount = ratingCount
	}
	return nil
}

func TestCatalogUsecase(t *testing.T) {
	repo := newFakeCatalogRepo()
	repo.categories = append(repo.categories, domaincatalog.Category{ID: "c-1", Name: "Laptops", Slug: "laptops"})
	repo.products["p-1"] = &domaincatalog.Product{
		ID:         "p-1",
		CategoryID: "c-1",
		Name:       "MacBook Pro 16",
		Slug:       "macbook-pro-16",
		Price:      money.VND(50000000),
		Stock:      10,
		Status:     domaincatalog.StatusActive,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	uc := New(repo)
	ctx := context.Background()

	cats, err := uc.ListCategories(ctx)
	if err != nil || len(cats) != 1 {
		t.Fatalf("failed to list categories: %v", err)
	}

	prods, total, err := uc.ListProducts(ctx, port.ProductFilter{Query: "MacBook"})
	if err != nil || total != 1 || len(prods) != 1 {
		t.Fatalf("failed to list products: %v, total: %d", err, total)
	}

	prod, err := uc.GetProductBySlug(ctx, "macbook-pro-16")
	if err != nil || prod.ID != "p-1" {
		t.Fatalf("failed to get product by slug: %v", err)
	}

	_, err = uc.GetProductBySlug(ctx, "non-existent")
	if err == nil {
		t.Fatalf("expected error for non-existent slug")
	}
}
