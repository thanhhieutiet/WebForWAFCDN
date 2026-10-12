package catalog

import (
	"strings"
	"time"

	"labshop/internal/domain/errs"
	"labshop/internal/domain/money"
)

type ProductStatus string

const (
	StatusActive   ProductStatus = "active"
	StatusArchived ProductStatus = "archived"
)

// Category represents a product category.
type Category struct {
	ID          string
	Name        string
	Slug        string
	Description string
	CreatedAt   time.Time
}

// Validate checks business rules for a category.
func (c *Category) Validate() error {
	var violations []errs.FieldViolation
	if strings.TrimSpace(c.Name) == "" {
		violations = append(violations, errs.FieldViolation{Field: "name", Issue: "category name is required"})
	}
	if strings.TrimSpace(c.Slug) == "" {
		violations = append(violations, errs.FieldViolation{Field: "slug", Issue: "category slug is required"})
	}
	if len(violations) > 0 {
		return errs.NewValidation("category validation failed", violations...)
	}
	return nil
}

// Product represents an item in the catalog.
type Product struct {
	ID          string
	CategoryID  string
	Name        string
	Slug        string
	Description string
	Price       money.VND
	Stock       int
	ImagePath   string
	Status      ProductStatus
	RatingAvg   float64
	RatingCount int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate checks product invariants.
func (p *Product) Validate() error {
	var violations []errs.FieldViolation

	if strings.TrimSpace(p.Name) == "" {
		violations = append(violations, errs.FieldViolation{Field: "name", Issue: "product name is required"})
	}
	if strings.TrimSpace(p.Slug) == "" {
		violations = append(violations, errs.FieldViolation{Field: "slug", Issue: "product slug is required"})
	}
	if strings.TrimSpace(p.CategoryID) == "" {
		violations = append(violations, errs.FieldViolation{Field: "category_id", Issue: "category_id is required"})
	}
	if p.Price <= 0 {
		violations = append(violations, errs.FieldViolation{Field: "price", Issue: "price must be strictly greater than 0 VND"})
	}
	if p.Stock < 0 {
		violations = append(violations, errs.FieldViolation{Field: "stock", Issue: "stock cannot be negative"})
	}
	if p.Status != StatusActive && p.Status != StatusArchived {
		violations = append(violations, errs.FieldViolation{Field: "status", Issue: "status must be active or archived"})
	}
	if p.RatingAvg < 0 || p.RatingAvg > 5.0 {
		violations = append(violations, errs.FieldViolation{Field: "rating_avg", Issue: "rating_avg must be between 0 and 5"})
	}
	if p.RatingCount < 0 {
		violations = append(violations, errs.FieldViolation{Field: "rating_count", Issue: "rating_count cannot be negative"})
	}

	if len(violations) > 0 {
		return errs.NewValidation("product validation failed", violations...)
	}
	return nil
}

// IsAvailable returns true if the product is active and has available stock.
func (p *Product) IsAvailable() bool {
	return p.Status == StatusActive && p.Stock > 0
}

// Archive changes product status to archived.
func (p *Product) Archive() {
	p.Status = StatusArchived
}
