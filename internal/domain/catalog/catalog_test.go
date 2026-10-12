package catalog

import (
	"testing"
	"time"

	"labshop/internal/domain/money"
)

func TestCategoryValidation(t *testing.T) {
	c := &Category{
		ID:        "c-1",
		Name:      "Smartphones",
		Slug:      "smartphones",
		CreatedAt: time.Now(),
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("expected category to be valid: %v", err)
	}

	c.Name = ""
	if err := c.Validate(); err == nil {
		t.Fatalf("expected validation error for empty name")
	}
}

func TestProductValidation(t *testing.T) {
	p := &Product{
		ID:          "p-1",
		CategoryID:  "c-1",
		Name:        "Flagship Phone",
		Slug:        "flagship-phone",
		Price:       money.VND(15000000),
		Stock:       10,
		Status:      StatusActive,
		RatingAvg:   4.5,
		RatingCount: 10,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("expected product to be valid: %v", err)
	}

	if !p.IsAvailable() {
		t.Fatalf("expected product to be available")
	}

	p.Price = money.VND(0)
	if err := p.Validate(); err == nil {
		t.Fatalf("expected validation error for zero price")
	}

	p.Price = money.VND(1000)
	p.Stock = -1
	if err := p.Validate(); err == nil {
		t.Fatalf("expected validation error for negative stock")
	}

	p.Stock = 5
	p.Archive()
	if p.Status != StatusArchived || p.IsAvailable() {
		t.Fatalf("archived product should not be available")
	}
}
