package coupon

import (
	"testing"
	"time"

	"labshop/internal/domain/money"
)

func TestCouponRules(t *testing.T) {
	now := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)

	// Invariant validation
	cBad := &Coupon{Code: "", Type: "invalid"}
	if err := cBad.ValidateInvariants(); err == nil {
		t.Errorf("expected invariant validation error")
	}

	cPercentBad := &Coupon{Code: "OVER50", Type: TypePercent, Value: 60}
	if err := cPercentBad.ValidateInvariants(); err == nil {
		t.Errorf("expected error for percent > 50")
	}

	// Percent coupon with cap
	cPercent := &Coupon{
		Code:        "WELCOME10",
		Type:        TypePercent,
		Value:       10,
		MaxDiscount: money.VND(100000),
		MinSubtotal: money.VND(200000),
		ValidFrom:   now.Add(-time.Hour),
		ValidTo:     now.Add(time.Hour),
		MaxUses:     10,
		UsedCount:   2,
	}

	// Below min subtotal
	_, _, err := cPercent.CalculateDiscount(money.VND(150000), now)
	if err == nil {
		t.Errorf("expected error for subtotal below minimum")
	}

	// 10% of 300,000 = 30,000 (below cap)
	d, freeShip, err := cPercent.CalculateDiscount(money.VND(300000), now)
	if err != nil || d != money.VND(30000) || freeShip {
		t.Errorf("expected 30,000 discount, got %v (err: %v)", d, err)
	}

	// 10% of 1,500,000 = 150,000 -> capped at 100,000
	d, _, err = cPercent.CalculateDiscount(money.VND(1500000), now)
	if err != nil || d != money.VND(100000) {
		t.Errorf("expected 100,000 discount capped, got %v", d)
	}

	// Expired coupon
	cPercent.ValidTo = now.Add(-time.Minute)
	_, _, err = cPercent.CalculateDiscount(money.VND(300000), now)
	if err == nil {
		t.Errorf("expected error for expired coupon")
	}

	// FreeShip coupon
	cShip := &Coupon{
		Code: "FREESHIP",
		Type: TypeFreeShip,
	}
	d, freeShip, err = cShip.CalculateDiscount(money.VND(100000), now)
	if err != nil || !freeShip || d != 0 {
		t.Errorf("expected freeship coupon to give freeShip=true and d=0, got %v, %v", freeShip, d)
	}

	// Fixed coupon
	cFixed := &Coupon{
		Code:  "FIXED50",
		Type:  TypeFixed,
		Value: 50000,
	}
	if err := cFixed.ValidateInvariants(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	dFixed, _, err := cFixed.CalculateDiscount(money.VND(100000), now)
	if err != nil || dFixed != money.VND(50000) {
		t.Fatalf("expected 50,000 fixed discount, got %v", dFixed)
	}
	// Fixed coupon exceeding subtotal
	dFixedOver, _, _ := cFixed.CalculateDiscount(money.VND(30000), now)
	if dFixedOver != money.VND(30000) {
		t.Fatalf("expected discount capped at subtotal, got %v", dFixedOver)
	}

	// Max uses reached
	cLimit := &Coupon{
		Code:      "LIMIT",
		Type:      TypeFixed,
		Value:     10000,
		MaxUses:   5,
		UsedCount: 5,
	}
	_, _, err = cLimit.CalculateDiscount(money.VND(50000), now)
	if err == nil {
		t.Fatalf("expected error when coupon limit reached")
	}

	// Not yet active
	cFuture := &Coupon{
		Code:      "FUTURE",
		Type:      TypeFixed,
		Value:     10000,
		ValidFrom: now.Add(time.Hour),
	}
	_, _, err = cFuture.CalculateDiscount(money.VND(50000), now)
	if err == nil {
		t.Fatalf("expected error for future coupon")
	}

	// Invalid type
	cUnknown := &Coupon{Code: "UNKNOWN", Type: "other"}
	_, _, err = cUnknown.CalculateDiscount(money.VND(50000), now)
	if err == nil {
		t.Fatalf("expected error for unknown coupon type")
	}
}
