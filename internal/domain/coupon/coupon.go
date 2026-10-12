package coupon

import (
	"fmt"
	"strings"
	"time"

	"labshop/internal/domain/errs"
	"labshop/internal/domain/money"
)

type Type string

const (
	TypePercent  Type = "percent"
	TypeFixed    Type = "fixed"
	TypeFreeShip Type = "freeship"
)

// Coupon represents a discount voucher.
type Coupon struct {
	ID          string
	Code        string
	Type        Type
	Value       int64 // Percentage (1..50) or fixed amount in VND
	MaxDiscount money.VND
	MinSubtotal money.VND
	ValidFrom   time.Time
	ValidTo     time.Time
	MaxUses     int
	UsedCount   int
	CreatedAt   time.Time
}

// ValidateInvariants validates the static configuration of the coupon.
func (c *Coupon) ValidateInvariants() error {
	var violations []errs.FieldViolation

	code := strings.TrimSpace(c.Code)
	if code == "" {
		violations = append(violations, errs.FieldViolation{Field: "code", Issue: "coupon code is required"})
	}

	switch c.Type {
	case TypePercent:
		if c.Value < 1 || c.Value > 50 {
			violations = append(violations, errs.FieldViolation{Field: "value", Issue: "percentage discount must be between 1 and 50"})
		}
	case TypeFixed:
		if c.Value <= 0 {
			violations = append(violations, errs.FieldViolation{Field: "value", Issue: "fixed discount must be greater than 0"})
		}
	case TypeFreeShip:
		// Free shipping coupon does not require a value
	default:
		violations = append(violations, errs.FieldViolation{Field: "type", Issue: "type must be percent, fixed, or freeship"})
	}

	if !c.ValidTo.IsZero() && !c.ValidFrom.IsZero() && c.ValidTo.Before(c.ValidFrom) {
		violations = append(violations, errs.FieldViolation{Field: "valid_to", Issue: "valid_to cannot be before valid_from"})
	}

	if len(violations) > 0 {
		return errs.NewValidation("coupon invariants invalid", violations...)
	}
	return nil
}

// CalculateDiscount validates rules against subtotal and time, and computes discount and whether freeship applies.
func (c *Coupon) CalculateDiscount(subtotal money.VND, now time.Time) (discount money.VND, freeShip bool, err error) {
	if !c.ValidFrom.IsZero() && now.Before(c.ValidFrom) {
		return 0, false, errs.NewBusinessRule("coupon_not_active", "coupon is not yet active")
	}
	if !c.ValidTo.IsZero() && now.After(c.ValidTo) {
		return 0, false, errs.NewBusinessRule("coupon_expired", "coupon has expired")
	}
	if c.MaxUses > 0 && c.UsedCount >= c.MaxUses {
		return 0, false, errs.NewBusinessRule("coupon_limit_reached", "coupon maximum usage limit reached")
	}
	if c.MinSubtotal > 0 && subtotal < c.MinSubtotal {
		return 0, false, errs.NewBusinessRule("coupon_min_subtotal", fmt.Sprintf("order subtotal must be at least %s", c.MinSubtotal.Format()))
	}

	switch c.Type {
	case TypePercent:
		calc := money.VND(subtotal.Int64() * c.Value / 100)
		if c.MaxDiscount > 0 && calc > c.MaxDiscount {
			calc = c.MaxDiscount
		}
		if calc > subtotal {
			calc = subtotal
		}
		return calc, false, nil

	case TypeFixed:
		fixedAmount := money.VND(c.Value)
		if fixedAmount > subtotal {
			fixedAmount = subtotal
		}
		return fixedAmount, false, nil

	case TypeFreeShip:
		return 0, true, nil

	default:
		return 0, false, errs.NewBusinessRule("invalid_coupon_type", "unsupported coupon type")
	}
}
