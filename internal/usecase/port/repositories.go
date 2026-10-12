package port

import (
	"context"

	"labshop/internal/domain/cart"
	"labshop/internal/domain/coupon"
	"labshop/internal/domain/order"
	"labshop/internal/domain/review"
	"labshop/internal/domain/support"
)

// CartRepository manages cart persistence per user.
type CartRepository interface {
	GetByUserID(ctx context.Context, userID string) (*cart.Cart, error)
	Save(ctx context.Context, c *cart.Cart) error
	DeleteByUserID(ctx context.Context, userID string) error
}

// CouponRepository manages coupon codes and their usages.
type CouponRepository interface {
	GetByCode(ctx context.Context, code string) (*coupon.Coupon, error)
	IncrementUsage(ctx context.Context, code string) error
	List(ctx context.Context) ([]coupon.Coupon, error)
	Save(ctx context.Context, c *coupon.Coupon) error
}

// OrderRepository manages orders and order items.
type OrderRepository interface {
	Create(ctx context.Context, o *order.Order) error
	GetByID(ctx context.Context, id string) (*order.Order, error)
	GetByIdempotencyKey(ctx context.Context, userID, key string) (*order.Order, error)
	UpdateStatus(ctx context.Context, id string, status order.Status, paymentRef string) error
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]order.Order, int, error)
	ListAll(ctx context.Context, limit, offset int) ([]order.Order, int, error)
	HasDeliveredOrderWithProduct(ctx context.Context, userID, productID string) (bool, error)
}

// ReviewRepository manages reviews and rating aggregation.
type ReviewRepository interface {
	Create(ctx context.Context, r *review.Review) error
	GetByID(ctx context.Context, id string) (*review.Review, error)
	GetByUserAndProduct(ctx context.Context, userID, productID string) (*review.Review, error)
	ListByProduct(ctx context.Context, productID string, limit, offset int) ([]review.Review, int, error)
	Delete(ctx context.Context, id string) error
	GetProductRatingStats(ctx context.Context, productID string) (avg float64, count int, err error)
}

// SupportRepository manages customer support tickets, messages, and attachments.
type SupportRepository interface {
	CreateTicket(ctx context.Context, t *support.Ticket) error
	GetTicketByID(ctx context.Context, id string) (*support.Ticket, error)
	ListTickets(ctx context.Context, userID string, limit, offset int) ([]support.Ticket, int, error)
	AddMessage(ctx context.Context, msg *support.TicketMessage) error
	AddAttachment(ctx context.Context, att *support.TicketAttachment) error
	UpdateTicketStatus(ctx context.Context, ticketID string, status support.TicketStatus) error
}
