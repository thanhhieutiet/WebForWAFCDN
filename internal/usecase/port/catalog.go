package port

import (
	"context"

	"labshop/internal/domain/catalog"
	"labshop/internal/domain/money"
	"labshop/internal/domain/user"
)

// UserRepository defines storage operations for users.
type UserRepository interface {
	Create(ctx context.Context, u *user.User) error
	GetByID(ctx context.Context, id string) (*user.User, error)
	GetByEmail(ctx context.Context, email string) (*user.User, error)
	Update(ctx context.Context, u *user.User) error
	List(ctx context.Context, limit, offset int) ([]user.User, int, error)
}

// ProductFilter parameters for catalog searching and filtering.
type ProductFilter struct {
	Query      string
	CategoryID string
	MinPrice   *money.VND
	MaxPrice   *money.VND
	Sort       string // "price_asc", "price_desc", "newest", "rating"
	Page       int
	PageSize   int
}

// CatalogRepository defines storage operations for catalog entities.
type CatalogRepository interface {
	ListCategories(ctx context.Context) ([]catalog.Category, error)
	GetCategoryByID(ctx context.Context, id string) (*catalog.Category, error)
	CreateCategory(ctx context.Context, c *catalog.Category) error

	ListProducts(ctx context.Context, filter ProductFilter) ([]catalog.Product, int, error)
	GetProductByID(ctx context.Context, id string) (*catalog.Product, error)
	GetProductBySlug(ctx context.Context, slug string) (*catalog.Product, error)
	CreateProduct(ctx context.Context, p *catalog.Product) error
	UpdateProduct(ctx context.Context, p *catalog.Product) error
	DeductStockIfAvailable(ctx context.Context, id string, quantity int) (bool, error)
	RestoreStock(ctx context.Context, id string, quantity int) error
	UpdateProductRating(ctx context.Context, id string, ratingAvg float64, ratingCount int) error
}
