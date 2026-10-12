package sqlite

import (
	"context"
	"fmt"

	"labshop/internal/domain/money"
	"labshop/internal/usecase/port"
)

type AdminRepository struct {
	db *DB
}

var _ port.AdminRepository = (*AdminRepository)(nil)

func NewAdminRepository(db *DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) GetStats(ctx context.Context) (*port.AdminStats, error) {
	stats := &port.AdminStats{
		OrdersByStatus: make(map[string]int),
	}

	// Total users
	_ = r.db.exec(ctx).QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)

	// Total orders
	_ = r.db.exec(ctx).QueryRowContext(ctx, "SELECT COUNT(*) FROM orders").Scan(&stats.TotalOrders)

	// Total revenue
	var totalRev int64
	_ = r.db.exec(ctx).QueryRowContext(ctx, "SELECT COALESCE(SUM(total), 0) FROM orders WHERE status IN ('PAID', 'SHIPPED', 'DELIVERED')").Scan(&totalRev)
	stats.TotalRevenueVND = money.VND(totalRev)

	// Orders by status
	rowsStatus, err := r.db.exec(ctx).QueryContext(ctx, "SELECT status, COUNT(*) FROM orders GROUP BY status")
	if err == nil {
		defer rowsStatus.Close()
		for rowsStatus.Next() {
			var (
				status string
				cnt    int
			)
			if err := rowsStatus.Scan(&status, &cnt); err == nil {
				stats.OrdersByStatus[status] = cnt
			}
		}
	}

	// Revenue by day
	queryRevByDay := `
		SELECT substr(created_at, 1, 10) as day, COALESCE(SUM(total), 0) as rev, COUNT(*) as cnt
		FROM orders
		WHERE status IN ('PAID', 'SHIPPED', 'DELIVERED')
		GROUP BY day
		ORDER BY day DESC
		LIMIT 30
	`
	rowsDays, err := r.db.exec(ctx).QueryContext(ctx, queryRevByDay)
	if err == nil {
		defer rowsDays.Close()
		for rowsDays.Next() {
			var (
				day string
				rev int64
				cnt int
			)
			if err := rowsDays.Scan(&day, &rev, &cnt); err == nil {
				stats.RevenueByDay = append(stats.RevenueByDay, port.DayRevenue{
					Date:       day,
					Revenue:    money.VND(rev),
					OrderCount: cnt,
				})
			}
		}
	}

	// Top products
	queryTopProds := `
		SELECT oi.product_id, oi.product_name, COALESCE(SUM(oi.quantity), 0) as sold, COALESCE(SUM(oi.subtotal), 0) as rev
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.status IN ('PAID', 'SHIPPED', 'DELIVERED')
		GROUP BY oi.product_id, oi.product_name
		ORDER BY sold DESC
		LIMIT 10
	`
	rowsTop, err := r.db.exec(ctx).QueryContext(ctx, queryTopProds)
	if err != nil {
		return nil, fmt.Errorf("failed to query top products: %w", err)
	}
	defer rowsTop.Close()

	for rowsTop.Next() {
		var (
			pid  string
			name string
			sold int
			rev  int64
		)
		if err := rowsTop.Scan(&pid, &name, &sold, &rev); err == nil {
			stats.TopProducts = append(stats.TopProducts, port.TopProduct{
				ProductID:    pid,
				ProductName:  name,
				QuantitySold: sold,
				Revenue:      money.VND(rev),
			})
		}
	}

	return stats, nil
}
