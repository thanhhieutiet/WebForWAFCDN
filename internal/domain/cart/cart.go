package cart

import (
	"time"

	"labshop/internal/domain/catalog"
	"labshop/internal/domain/errs"
)

const (
	MaxLinesPerCart = 20
	MinItemQuantity = 1
	MaxItemQuantity = 10
)

// CartItem represents an item entry in a user's shopping cart.
type CartItem struct {
	ProductID string
	Quantity  int
	UpdatedAt time.Time
}

// Cart represents a user's persistent cart.
type Cart struct {
	UserID     string
	Items      []CartItem
	CouponCode string
	UpdatedAt  time.Time
}

// NewCart creates an empty cart for a given user.
func NewCart(userID string) *Cart {
	return &Cart{
		UserID:    userID,
		Items:     make([]CartItem, 0),
		UpdatedAt: time.Now(),
	}
}

// SetItem updates or adds an item with the given quantity, validating invariants against the product.
func (c *Cart) SetItem(prod *catalog.Product, quantity int) error {
	if quantity < MinItemQuantity || quantity > MaxItemQuantity {
		return errs.NewValidation("quantity must be between 1 and 10", errs.FieldViolation{
			Field: "quantity",
			Issue: "must be between 1 and 10",
		})
	}

	if prod.Status == catalog.StatusArchived {
		return errs.NewBusinessRule("product_archived", "cannot add archived product to cart")
	}

	if quantity > prod.Stock {
		return errs.NewConflict("insufficient stock for product")
	}

	for i := range c.Items {
		if c.Items[i].ProductID == prod.ID {
			c.Items[i].Quantity = quantity
			c.Items[i].UpdatedAt = time.Now()
			c.UpdatedAt = time.Now()
			return nil
		}
	}

	if len(c.Items) >= MaxLinesPerCart {
		return errs.NewBusinessRule("cart_full", "cart cannot exceed 20 items")
	}

	c.Items = append(c.Items, CartItem{
		ProductID: prod.ID,
		Quantity:  quantity,
		UpdatedAt: time.Now(),
	})
	c.UpdatedAt = time.Now()
	return nil
}

// RemoveItem removes a product from the cart.
func (c *Cart) RemoveItem(productID string) bool {
	for i, item := range c.Items {
		if item.ProductID == productID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			c.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// Clear empties all items and resets coupon in the cart.
func (c *Cart) Clear() {
	c.Items = make([]CartItem, 0)
	c.CouponCode = ""
	c.UpdatedAt = time.Now()
}

// ApplyCoupon sets the coupon code (exactly 1 coupon per cart).
func (c *Cart) ApplyCoupon(code string) {
	c.CouponCode = code
	c.UpdatedAt = time.Now()
}

// RemoveCoupon removes the applied coupon.
func (c *Cart) RemoveCoupon() {
	c.CouponCode = ""
	c.UpdatedAt = time.Now()
}

// FindItem searches for an item in the cart by product ID.
func (c *Cart) FindItem(productID string) *CartItem {
	for i := range c.Items {
		if c.Items[i].ProductID == productID {
			return &c.Items[i]
		}
	}
	return nil
}
