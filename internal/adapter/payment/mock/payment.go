package mock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	mrand "math/rand"
	"time"

	"labshop/internal/domain/money"
)

type Gateway struct {
	failRate float64 // between 0.0 and 1.0
}

func New(failRate float64) *Gateway {
	if failRate > 1.0 {
		failRate = failRate / 100.0 // Normalize 0..100 to 0..1
	}
	if failRate < 0 {
		failRate = 0
	}
	return &Gateway{failRate: failRate}
}

func (g *Gateway) Charge(ctx context.Context, orderID string, amount money.VND) (string, error) {
	// 200 - 800ms simulated delay
	delayMs := 200 + mrand.Intn(601)
	select {
	case <-time.After(time.Duration(delayMs) * time.Millisecond):
	case <-ctx.Done():
		return "", ctx.Err()
	}

	if g.failRate > 0 && mrand.Float64() < g.failRate {
		return "", errors.New("mock payment processor declined transaction")
	}

	randBytes := make([]byte, 8)
	_, _ = rand.Read(randBytes)
	ref := fmt.Sprintf("mock_pay_%s_%s", orderID, hex.EncodeToString(randBytes))

	return ref, nil
}
