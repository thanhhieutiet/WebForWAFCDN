package cart

import (
	"fmt"
	"testing"
	"time"

	"labshop/internal/domain/catalog"
	"labshop/internal/domain/money"
)

func TestCartOperations(t *testing.T) {
	c := NewCart("u-1")
	prod := &catalog.Product{
		ID:        "p-1",
		Price:     money.VND(100000),
		Stock:     5,
		Status:    catalog.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Quantity bounds
	if err := c.SetItem(prod, 0); err == nil {
		t.Errorf("expected error for quantity 0")
	}
	if err := c.SetItem(prod, 11); err == nil {
		t.Errorf("expected error for quantity 11")
	}

	// Exceed stock
	if err := c.SetItem(prod, 6); err == nil {
		t.Errorf("expected conflict error when quantity exceeds stock")
	}

	// Valid set
	if err := c.SetItem(prod, 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Items) != 1 || c.Items[0].Quantity != 3 {
		t.Fatalf("item not added correctly")
	}

	// Update existing item
	if err := c.SetItem(prod, 5); err != nil {
		t.Fatalf("unexpected error on update: %v", err)
	}
	if len(c.Items) != 1 || c.Items[0].Quantity != 5 {
		t.Fatalf("item not updated correctly")
	}

	// Archived product
	archivedProd := &catalog.Product{
		ID:     "p-archived",
		Price:  money.VND(50000),
		Stock:  10,
		Status: catalog.StatusArchived,
	}
	if err := c.SetItem(archivedProd, 1); err == nil {
		t.Errorf("expected error when adding archived product")
	}

	// Max 20 lines
	for i := 2; i <= 20; i++ {
		p := &catalog.Product{
			ID:     fmt.Sprintf("p-%d", i),
			Stock:  10,
			Status: catalog.StatusActive,
		}
		if err := c.SetItem(p, 1); err != nil {
			t.Fatalf("failed to add item %d: %v", i, err)
		}
	}
	// 21st item should fail
	extraProd := &catalog.Product{ID: "p-extra", Stock: 10, Status: catalog.StatusActive}
	if err := c.SetItem(extraProd, 1); err == nil {
		t.Errorf("expected error exceeding 20 items")
	}

	// Remove item
	if !c.RemoveItem("p-1") {
		t.Errorf("expected true when removing p-1")
	}
	if c.FindItem("p-1") != nil {
		t.Errorf("expected p-1 to be removed")
	}

	// Coupon
	c.ApplyCoupon("DISCOUNT10")
	if c.CouponCode != "DISCOUNT10" {
		t.Errorf("expected coupon DISCOUNT10")
	}
	c.RemoveCoupon()
	if c.CouponCode != "" {
		t.Errorf("expected coupon to be cleared")
	}

	// Clear
	c.Clear()
	if len(c.Items) != 0 {
		t.Errorf("expected empty cart after clear")
	}
}
