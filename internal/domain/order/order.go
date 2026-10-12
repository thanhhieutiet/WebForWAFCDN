package order

import (
	"fmt"
	"strings"
	"time"

	"labshop/internal/domain/errs"
	"labshop/internal/domain/money"
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusPaid      Status = "PAID"
	StatusShipped   Status = "SHIPPED"
	StatusDelivered Status = "DELIVERED"
	StatusCancelled Status = "CANCELLED"
)

const (
	DefaultShippingFee   money.VND = 30000
	FreeShippingMinOrder money.VND = 500000
)

const (
	PaymentMethodMockGateway = "mock_gateway"
	PaymentMethodCOD         = "cod"
)

// AllowedTransitions maps each valid initial state to allowed destination states.
var AllowedTransitions = map[Status][]Status{
	StatusPending:   {StatusPaid, StatusCancelled},
	StatusPaid:      {StatusShipped, StatusCancelled},
	StatusShipped:   {StatusDelivered},
	StatusDelivered: {},
	StatusCancelled: {},
}

// CanTransition returns true if the transition from currentState to targetState is valid.
func CanTransition(from, to Status) bool {
	targets, ok := AllowedTransitions[from]
	if !ok {
		return false
	}
	for _, target := range targets {
		if target == to {
			return true
		}
	}
	return false
}

// CalculateShipping determines shipping fee based on subtotal and coupon flag.
func CalculateShipping(subtotal money.VND, freeShipCoupon bool) money.VND {
	if freeShipCoupon || subtotal >= FreeShippingMinOrder {
		return 0
	}
	return DefaultShippingFee
}

// OrderItem represents an individual line item in an order.
type OrderItem struct {
	ID           string
	OrderID      string
	ProductID    string
	ProductName  string
	ProductPrice money.VND
	Quantity     int
	Subtotal     money.VND
}

// Order represents a customer purchase order.
type Order struct {
	ID              string
	UserID          string
	Items           []OrderItem
	Subtotal        money.VND
	Discount        money.VND
	ShippingFee     money.VND
	Total           money.VND
	CouponCode      string
	Status          Status
	PaymentMethod   string
	PaymentRef      string
	RecipientName   string
	RecipientPhone  string
	ShippingAddress string
	IdempotencyKey  string
	PayloadHash     string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ValidateInvariants ensures the order is in a valid state when created.
func (o *Order) ValidateInvariants() error {
	var violations []errs.FieldViolation

	if strings.TrimSpace(o.UserID) == "" {
		violations = append(violations, errs.FieldViolation{Field: "user_id", Issue: "user_id is required"})
	}
	if len(o.Items) == 0 {
		violations = append(violations, errs.FieldViolation{Field: "items", Issue: "order must contain at least one item"})
	}
	if o.PaymentMethod != PaymentMethodMockGateway && o.PaymentMethod != PaymentMethodCOD {
		violations = append(violations, errs.FieldViolation{Field: "payment_method", Issue: "payment_method must be mock_gateway or cod"})
	}
	if strings.TrimSpace(o.RecipientName) == "" {
		violations = append(violations, errs.FieldViolation{Field: "recipient_name", Issue: "recipient name is required"})
	}
	if strings.TrimSpace(o.RecipientPhone) == "" {
		violations = append(violations, errs.FieldViolation{Field: "recipient_phone", Issue: "recipient phone is required"})
	}
	if strings.TrimSpace(o.ShippingAddress) == "" {
		violations = append(violations, errs.FieldViolation{Field: "shipping_address", Issue: "shipping address is required"})
	}

	if len(violations) > 0 {
		return errs.NewValidation("order validation failed", violations...)
	}
	return nil
}

// TransitionTo attempts to transition the order to the given status.
func (o *Order) TransitionTo(target Status) error {
	if !CanTransition(o.Status, target) {
		return errs.NewBusinessRule("invalid_order_state_transition", fmt.Sprintf("cannot transition order from %s to %s", o.Status, target))
	}
	o.Status = target
	o.UpdatedAt = time.Now()
	return nil
}
