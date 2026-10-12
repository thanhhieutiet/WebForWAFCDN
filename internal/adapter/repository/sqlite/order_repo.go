package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"labshop/internal/domain/money"
	"labshop/internal/domain/order"
	"labshop/internal/usecase/port"
)

type OrderRepository struct {
	db *DB
}

var _ port.OrderRepository = (*OrderRepository)(nil)

func NewOrderRepository(db *DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, o *order.Order) error {
	queryOrder := `
		INSERT INTO orders (
			id, user_id, subtotal, discount, shipping_fee, total,
			coupon_code, status, payment_method, payment_ref,
			recipient_name, recipient_phone, shipping_address,
			idempotency_key, payload_hash, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.exec(ctx).ExecContext(ctx, queryOrder,
		o.ID, o.UserID, o.Subtotal.Int64(), o.Discount.Int64(), o.ShippingFee.Int64(), o.Total.Int64(),
		o.CouponCode, string(o.Status), o.PaymentMethod, o.PaymentRef,
		o.RecipientName, o.RecipientPhone, o.ShippingAddress,
		o.IdempotencyKey, o.PayloadHash,
		o.CreatedAt.UTC().Format(time.RFC3339),
		o.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	queryItem := `
		INSERT INTO order_items (id, order_id, product_id, product_name, product_price, quantity, subtotal)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	for _, item := range o.Items {
		_, err := r.db.exec(ctx).ExecContext(ctx, queryItem,
			item.ID, o.ID, item.ProductID, item.ProductName, item.ProductPrice.Int64(), item.Quantity, item.Subtotal.Int64(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	return nil
}

func (r *OrderRepository) loadOrderItems(ctx context.Context, orderID string) ([]order.OrderItem, error) {
	rows, err := r.db.exec(ctx).QueryContext(ctx,
		"SELECT id, order_id, product_id, product_name, product_price, quantity, subtotal FROM order_items WHERE order_id = ?",
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []order.OrderItem
	for rows.Next() {
		var (
			it       order.OrderItem
			price    int64
			subtotal int64
		)
		if err := rows.Scan(&it.ID, &it.OrderID, &it.ProductID, &it.ProductName, &price, &it.Quantity, &subtotal); err != nil {
			return nil, err
		}
		it.ProductPrice = money.VND(price)
		it.Subtotal = money.VND(subtotal)
		items = append(items, it)
	}
	return items, nil
}

func (r *OrderRepository) scanOrder(ctx context.Context, row *sql.Row) (*order.Order, error) {
	var (
		o           order.Order
		subtotal    int64
		discount    int64
		shippingFee int64
		total       int64
		statusStr   string
		createdAt   string
		updatedAt   string
	)

	err := row.Scan(
		&o.ID, &o.UserID, &subtotal, &discount, &shippingFee, &total,
		&o.CouponCode, &statusStr, &o.PaymentMethod, &o.PaymentRef,
		&o.RecipientName, &o.RecipientPhone, &o.ShippingAddress,
		&o.IdempotencyKey, &o.PayloadHash, &createdAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("order not found")
	}
	if err != nil {
		return nil, err
	}

	o.Subtotal = money.VND(subtotal)
	o.Discount = money.VND(discount)
	o.ShippingFee = money.VND(shippingFee)
	o.Total = money.VND(total)
	o.Status = order.Status(statusStr)
	o.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	o.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	items, err := r.loadOrderItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*order.Order, error) {
	query := `
		SELECT id, user_id, subtotal, discount, shipping_fee, total,
		       coupon_code, status, payment_method, payment_ref,
		       recipient_name, recipient_phone, shipping_address,
		       idempotency_key, payload_hash, created_at, updated_at
		FROM orders WHERE id = ?
	`
	row := r.db.exec(ctx).QueryRowContext(ctx, query, id)
	return r.scanOrder(ctx, row)
}

func (r *OrderRepository) GetByIdempotencyKey(ctx context.Context, userID, key string) (*order.Order, error) {
	query := `
		SELECT id, user_id, subtotal, discount, shipping_fee, total,
		       coupon_code, status, payment_method, payment_ref,
		       recipient_name, recipient_phone, shipping_address,
		       idempotency_key, payload_hash, created_at, updated_at
		FROM orders WHERE user_id = ? AND idempotency_key = ?
	`
	row := r.db.exec(ctx).QueryRowContext(ctx, query, userID, key)
	return r.scanOrder(ctx, row)
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status order.Status, paymentRef string) error {
	var query string
	var args []any
	nowStr := time.Now().UTC().Format(time.RFC3339)

	if paymentRef != "" {
		query = `UPDATE orders SET status = ?, payment_ref = ?, updated_at = ? WHERE id = ?`
		args = []any{string(status), paymentRef, nowStr, id}
	} else {
		query = `UPDATE orders SET status = ?, updated_at = ? WHERE id = ?`
		args = []any{string(status), nowStr, id}
	}

	_, err := r.db.exec(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	return nil
}

func (r *OrderRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]order.Order, int, error) {
	var total int
	err := r.db.exec(ctx).QueryRowContext(ctx, "SELECT COUNT(*) FROM orders WHERE user_id = ?", userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, subtotal, discount, shipping_fee, total,
		       coupon_code, status, payment_method, payment_ref,
		       recipient_name, recipient_phone, shipping_address,
		       idempotency_key, payload_hash, created_at, updated_at
		FROM orders
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.exec(ctx).QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []order.Order
	for rows.Next() {
		var (
			o           order.Order
			subtotal    int64
			discount    int64
			shippingFee int64
			totalVND    int64
			statusStr   string
			createdAt   string
			updatedAt   string
		)
		err := rows.Scan(
			&o.ID, &o.UserID, &subtotal, &discount, &shippingFee, &totalVND,
			&o.CouponCode, &statusStr, &o.PaymentMethod, &o.PaymentRef,
			&o.RecipientName, &o.RecipientPhone, &o.ShippingAddress,
			&o.IdempotencyKey, &o.PayloadHash, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		o.Subtotal = money.VND(subtotal)
		o.Discount = money.VND(discount)
		o.ShippingFee = money.VND(shippingFee)
		o.Total = money.VND(totalVND)
		o.Status = order.Status(statusStr)
		o.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		o.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		items, err := r.loadOrderItems(ctx, o.ID)
		if err == nil {
			o.Items = items
		}
		orders = append(orders, o)
	}

	return orders, total, nil
}

func (r *OrderRepository) ListAll(ctx context.Context, limit, offset int) ([]order.Order, int, error) {
	var total int
	err := r.db.exec(ctx).QueryRowContext(ctx, "SELECT COUNT(*) FROM orders").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, subtotal, discount, shipping_fee, total,
		       coupon_code, status, payment_method, payment_ref,
		       recipient_name, recipient_phone, shipping_address,
		       idempotency_key, payload_hash, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.exec(ctx).QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []order.Order
	for rows.Next() {
		var (
			o           order.Order
			subtotal    int64
			discount    int64
			shippingFee int64
			totalVND    int64
			statusStr   string
			createdAt   string
			updatedAt   string
		)
		err := rows.Scan(
			&o.ID, &o.UserID, &subtotal, &discount, &shippingFee, &totalVND,
			&o.CouponCode, &statusStr, &o.PaymentMethod, &o.PaymentRef,
			&o.RecipientName, &o.RecipientPhone, &o.ShippingAddress,
			&o.IdempotencyKey, &o.PayloadHash, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		o.Subtotal = money.VND(subtotal)
		o.Discount = money.VND(discount)
		o.ShippingFee = money.VND(shippingFee)
		o.Total = money.VND(totalVND)
		o.Status = order.Status(statusStr)
		o.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		o.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		items, err := r.loadOrderItems(ctx, o.ID)
		if err == nil {
			o.Items = items
		}
		orders = append(orders, o)
	}

	return orders, total, nil
}

func (r *OrderRepository) HasDeliveredOrderWithProduct(ctx context.Context, userID, productID string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM orders o
		JOIN order_items oi ON o.id = oi.order_id
		WHERE o.user_id = ? AND o.status = 'DELIVERED' AND oi.product_id = ?
	`
	var count int
	err := r.db.exec(ctx).QueryRowContext(ctx, query, userID, productID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
