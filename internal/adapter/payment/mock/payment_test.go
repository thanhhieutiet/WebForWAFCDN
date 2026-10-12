package mock

import (
	"context"
	"testing"
	"time"

	"labshop/internal/domain/money"
)

func TestMockPaymentGateway(t *testing.T) {
	gwSuccess := New(0.0)

	start := time.Now()
	ref, err := gwSuccess.Charge(context.Background(), "ord-1", money.VND(100000))
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref == "" {
		t.Fatalf("expected non-empty payment reference")
	}
	if duration < 190*time.Millisecond {
		t.Fatalf("expected delay >= 200ms, got %v", duration)
	}

	// 100% fail rate
	gwFail := New(1.0)
	_, err = gwFail.Charge(context.Background(), "ord-1", money.VND(100000))
	if err == nil {
		t.Fatalf("expected payment decline error for 100%% fail rate")
	}

	// Context cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err = gwSuccess.Charge(ctx, "ord-2", money.VND(50000))
	if err == nil {
		t.Fatalf("expected context timeout error")
	}
}
