package order

import (
	"testing"
	"time"

	"labshop/internal/domain/money"
)

func TestOrderStateMachine(t *testing.T) {
	o := &Order{
		ID:        "ord-1",
		UserID:    "u-1",
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// PENDING -> PAID
	if err := o.TransitionTo(StatusPaid); err != nil {
		t.Fatalf("unexpected transition error: %v", err)
	}
	if o.Status != StatusPaid {
		t.Fatalf("expected status PAID, got %s", o.Status)
	}

	// PAID -> DELIVERED (invalid, must be SHIPPED first)
	if err := o.TransitionTo(StatusDelivered); err == nil {
		t.Fatalf("expected error transitioning directly from PAID to DELIVERED")
	}

	// PAID -> SHIPPED
	if err := o.TransitionTo(StatusShipped); err != nil {
		t.Fatalf("unexpected transition error: %v", err)
	}

	// SHIPPED -> CANCELLED (invalid)
	if err := o.TransitionTo(StatusCancelled); err == nil {
		t.Fatalf("expected error cancelling already shipped order")
	}

	// SHIPPED -> DELIVERED
	if err := o.TransitionTo(StatusDelivered); err != nil {
		t.Fatalf("unexpected transition error: %v", err)
	}

	// DELIVERED -> anything (terminal)
	if err := o.TransitionTo(StatusCancelled); err == nil {
		t.Fatalf("expected error transitioning from terminal DELIVERED")
	}

	// PENDING -> CANCELLED
	oPending := &Order{Status: StatusPending}
	if err := oPending.TransitionTo(StatusCancelled); err != nil {
		t.Fatalf("expected PENDING to CANCELLED to be valid")
	}

	// CANCELLED -> anything (terminal)
	if err := oPending.TransitionTo(StatusPaid); err == nil {
		t.Fatalf("expected error transitioning from terminal CANCELLED")
	}
}

func TestShippingCalculation(t *testing.T) {
	// Subtotal < 500,000 and freeShip=false -> 30,000
	if fee := CalculateShipping(money.VND(200000), false); fee != 30000 {
		t.Fatalf("expected fee 30000, got %v", fee)
	}

	// Subtotal >= 500,000 -> 0
	if fee := CalculateShipping(money.VND(500000), false); fee != 0 {
		t.Fatalf("expected fee 0, got %v", fee)
	}

	// FreeShip coupon flag -> 0
	if fee := CalculateShipping(money.VND(100000), true); fee != 0 {
		t.Fatalf("expected fee 0 with freeShip coupon, got %v", fee)
	}
}

func TestOrderValidateInvariants(t *testing.T) {
	o := &Order{}
	if err := o.ValidateInvariants(); err == nil {
		t.Errorf("expected validation error for empty order")
	}

	o.UserID = "u-1"
	o.Items = []OrderItem{{ProductID: "p-1", Quantity: 1}}
	o.PaymentMethod = "invalid"
	o.RecipientName = "Name"
	o.RecipientPhone = "0900"
	o.ShippingAddress = "Address"
	if err := o.ValidateInvariants(); err == nil {
		t.Errorf("expected validation error for invalid payment method")
	}

	o.PaymentMethod = PaymentMethodMockGateway
	if err := o.ValidateInvariants(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

