package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"labshop/internal/domain/cart"
	"labshop/internal/usecase/port"
)

type CartRepository struct {
	db *DB
}

var _ port.CartRepository = (*CartRepository)(nil)

func NewCartRepository(db *DB) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) GetByUserID(ctx context.Context, userID string) (*cart.Cart, error) {
	var (
		couponCode string
		updatedAt  string
	)
	err := r.db.exec(ctx).QueryRowContext(ctx, "SELECT coupon_code, updated_at FROM carts WHERE user_id = ?", userID).
		Scan(&couponCode, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("cart not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query cart: %w", err)
	}

	c := &cart.Cart{
		UserID:     userID,
		CouponCode: couponCode,
		Items:      make([]cart.CartItem, 0),
	}
	c.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	// Fetch items
	rows, err := r.db.exec(ctx).QueryContext(ctx, "SELECT product_id, quantity, updated_at FROM cart_items WHERE user_id = ?", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cart items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			it           cart.CartItem
			itemUpdateAt string
		)
		if err := rows.Scan(&it.ProductID, &it.Quantity, &itemUpdateAt); err != nil {
			return nil, err
		}
		it.UpdatedAt, _ = time.Parse(time.RFC3339, itemUpdateAt)
		c.Items = append(c.Items, it)
	}

	return c, nil
}

func (r *CartRepository) Save(ctx context.Context, c *cart.Cart) error {
	nowStr := c.UpdatedAt.UTC().Format(time.RFC3339)

	// Upsert cart header
	queryCart := `
		INSERT INTO carts (user_id, coupon_code, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET coupon_code = excluded.coupon_code, updated_at = excluded.updated_at
	`
	if _, err := r.db.exec(ctx).ExecContext(ctx, queryCart, c.UserID, c.CouponCode, nowStr); err != nil {
		return fmt.Errorf("failed to upsert cart: %w", err)
	}

	// Remove old items
	if _, err := r.db.exec(ctx).ExecContext(ctx, "DELETE FROM cart_items WHERE user_id = ?", c.UserID); err != nil {
		return fmt.Errorf("failed to delete old cart items: %w", err)
	}

	// Insert items
	queryItem := `INSERT INTO cart_items (user_id, product_id, quantity, updated_at) VALUES (?, ?, ?, ?)`
	for _, it := range c.Items {
		itTimeStr := it.UpdatedAt.UTC().Format(time.RFC3339)
		if _, err := r.db.exec(ctx).ExecContext(ctx, queryItem, c.UserID, it.ProductID, it.Quantity, itTimeStr); err != nil {
			return fmt.Errorf("failed to insert cart item: %w", err)
		}
	}

	return nil
}

func (r *CartRepository) DeleteByUserID(ctx context.Context, userID string) error {
	_, _ = r.db.exec(ctx).ExecContext(ctx, "DELETE FROM cart_items WHERE user_id = ?", userID)
	_, err := r.db.exec(ctx).ExecContext(ctx, "DELETE FROM carts WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("failed to delete cart: %w", err)
	}
	return nil
}
