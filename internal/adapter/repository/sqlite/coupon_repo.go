package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"labshop/internal/domain/coupon"
	"labshop/internal/domain/money"
	"labshop/internal/usecase/port"
)

type CouponRepository struct {
	db *DB
}

var _ port.CouponRepository = (*CouponRepository)(nil)

func NewCouponRepository(db *DB) *CouponRepository {
	return &CouponRepository{db: db}
}

func (r *CouponRepository) GetByCode(ctx context.Context, code string) (*coupon.Coupon, error) {
	query := `
		SELECT id, code, type, value, max_discount, min_subtotal, valid_from, valid_to, max_uses, used_count, created_at
		FROM coupons WHERE UPPER(code) = UPPER(?)
	`
	var (
		c           coupon.Coupon
		typeStr     string
		maxDiscount int64
		minSubtotal int64
		validFrom   sql.NullString
		validTo     sql.NullString
		createdAt   string
	)
	err := r.db.exec(ctx).QueryRowContext(ctx, query, code).Scan(
		&c.ID, &c.Code, &typeStr, &c.Value, &maxDiscount, &minSubtotal,
		&validFrom, &validTo, &c.MaxUses, &c.UsedCount, &createdAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("coupon not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query coupon: %w", err)
	}

	c.Type = coupon.Type(typeStr)
	c.MaxDiscount = money.VND(maxDiscount)
	c.MinSubtotal = money.VND(minSubtotal)
	if validFrom.Valid && validFrom.String != "" {
		c.ValidFrom, _ = time.Parse(time.RFC3339, validFrom.String)
	}
	if validTo.Valid && validTo.String != "" {
		c.ValidTo, _ = time.Parse(time.RFC3339, validTo.String)
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

	return &c, nil
}

func (r *CouponRepository) IncrementUsage(ctx context.Context, code string) error {
	query := `UPDATE coupons SET used_count = used_count + 1 WHERE UPPER(code) = UPPER(?)`
	_, err := r.db.exec(ctx).ExecContext(ctx, query, code)
	if err != nil {
		return fmt.Errorf("failed to increment coupon usage: %w", err)
	}
	return nil
}

func (r *CouponRepository) List(ctx context.Context) ([]coupon.Coupon, error) {
	query := `
		SELECT id, code, type, value, max_discount, min_subtotal, valid_from, valid_to, max_uses, used_count, created_at
		FROM coupons ORDER BY created_at DESC
	`
	rows, err := r.db.exec(ctx).QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list coupons: %w", err)
	}
	defer rows.Close()

	var list []coupon.Coupon
	for rows.Next() {
		var (
			c           coupon.Coupon
			typeStr     string
			maxDiscount int64
			minSubtotal int64
			validFrom   sql.NullString
			validTo     sql.NullString
			createdAt   string
		)
		err := rows.Scan(
			&c.ID, &c.Code, &typeStr, &c.Value, &maxDiscount, &minSubtotal,
			&validFrom, &validTo, &c.MaxUses, &c.UsedCount, &createdAt,
		)
		if err != nil {
			return nil, err
		}
		c.Type = coupon.Type(typeStr)
		c.MaxDiscount = money.VND(maxDiscount)
		c.MinSubtotal = money.VND(minSubtotal)
		if validFrom.Valid && validFrom.String != "" {
			c.ValidFrom, _ = time.Parse(time.RFC3339, validFrom.String)
		}
		if validTo.Valid && validTo.String != "" {
			c.ValidTo, _ = time.Parse(time.RFC3339, validTo.String)
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		list = append(list, c)
	}
	return list, nil
}

func (r *CouponRepository) Save(ctx context.Context, c *coupon.Coupon) error {
	query := `
		INSERT INTO coupons (id, code, type, value, max_discount, min_subtotal, valid_from, valid_to, max_uses, used_count, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(code) DO UPDATE SET
			type = excluded.type,
			value = excluded.value,
			max_discount = excluded.max_discount,
			min_subtotal = excluded.min_subtotal,
			valid_from = excluded.valid_from,
			valid_to = excluded.valid_to,
			max_uses = excluded.max_uses,
			used_count = excluded.used_count
	`
	var validFromStr, validToStr sql.NullString
	if !c.ValidFrom.IsZero() {
		validFromStr = sql.NullString{String: c.ValidFrom.UTC().Format(time.RFC3339), Valid: true}
	}
	if !c.ValidTo.IsZero() {
		validToStr = sql.NullString{String: c.ValidTo.UTC().Format(time.RFC3339), Valid: true}
	}

	_, err := r.db.exec(ctx).ExecContext(ctx, query,
		c.ID, c.Code, string(c.Type), c.Value, c.MaxDiscount.Int64(), c.MinSubtotal.Int64(),
		validFromStr, validToStr, c.MaxUses, c.UsedCount, c.CreatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to save coupon: %w", err)
	}
	return nil
}
