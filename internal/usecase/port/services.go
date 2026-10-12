package port

import (
	"context"
	"time"

	"labshop/internal/domain/money"
	"labshop/internal/domain/user"
)

// TxManager manages database transactions.
type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// PasswordHasher abstracts password hashing and verification.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) bool
}

// Claims represents authenticated token payload.
type Claims struct {
	UserID    string
	Role      user.Role
	Email     string
	ExpiresAt time.Time
}

// TokenIssuer abstracts JWT generation and validation.
type TokenIssuer interface {
	IssueToken(u *user.User) (string, error)
	ValidateToken(tokenStr string) (*Claims, error)
}

// PaymentGateway simulates an external payment service.
type PaymentGateway interface {
	Charge(ctx context.Context, orderID string, amount money.VND) (ref string, err error)
}

// IDGenerator generates unique IDs (e.g. UUID).
type IDGenerator interface {
	NewID() string
}

// DayRevenue aggregates revenue per calendar day.
type DayRevenue struct {
	Date       string
	Revenue    money.VND
	OrderCount int
}

// TopProduct represents a best-selling item.
type TopProduct struct {
	ProductID    string
	ProductName  string
	QuantitySold int
	Revenue      money.VND
}

// AdminStats provides dashboard reporting metrics.
type AdminStats struct {
	RevenueByDay    []DayRevenue
	TopProducts     []TopProduct
	OrdersByStatus  map[string]int
	TotalUsers      int
	TotalOrders     int
	TotalRevenueVND money.VND
}

// AdminRepository provides administrative querying and reporting.
type AdminRepository interface {
	GetStats(ctx context.Context) (*AdminStats, error)
}
