package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"labshop/internal/domain/review"
	"labshop/internal/usecase/port"
)

type ReviewRepository struct {
	db *DB
}

var _ port.ReviewRepository = (*ReviewRepository)(nil)

func NewReviewRepository(db *DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Create(ctx context.Context, rev *review.Review) error {
	query := `
		INSERT INTO reviews (id, user_id, product_id, user_name, rating, comment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.exec(ctx).ExecContext(ctx, query,
		rev.ID, rev.UserID, rev.ProductID, rev.UserName, rev.Rating, rev.Comment,
		rev.CreatedAt.UTC().Format(time.RFC3339),
		rev.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to insert review: %w", err)
	}
	return nil
}

func (r *ReviewRepository) GetByID(ctx context.Context, id string) (*review.Review, error) {
	query := `
		SELECT id, user_id, product_id, user_name, rating, comment, created_at, updated_at
		FROM reviews WHERE id = ?
	`
	var (
		rev       review.Review
		createdAt string
		updatedAt string
	)
	err := r.db.exec(ctx).QueryRowContext(ctx, query, id).Scan(
		&rev.ID, &rev.UserID, &rev.ProductID, &rev.UserName, &rev.Rating, &rev.Comment, &createdAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("review not found")
	}
	if err != nil {
		return nil, err
	}
	rev.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	rev.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &rev, nil
}

func (r *ReviewRepository) GetByUserAndProduct(ctx context.Context, userID, productID string) (*review.Review, error) {
	query := `
		SELECT id, user_id, product_id, user_name, rating, comment, created_at, updated_at
		FROM reviews WHERE user_id = ? AND product_id = ?
	`
	var (
		rev       review.Review
		createdAt string
		updatedAt string
	)
	err := r.db.exec(ctx).QueryRowContext(ctx, query, userID, productID).Scan(
		&rev.ID, &rev.UserID, &rev.ProductID, &rev.UserName, &rev.Rating, &rev.Comment, &createdAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("review not found")
	}
	if err != nil {
		return nil, err
	}
	rev.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	rev.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &rev, nil
}

func (r *ReviewRepository) ListByProduct(ctx context.Context, productID string, limit, offset int) ([]review.Review, int, error) {
	var total int
	err := r.db.exec(ctx).QueryRowContext(ctx, "SELECT COUNT(*) FROM reviews WHERE product_id = ?", productID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, product_id, user_name, rating, comment, created_at, updated_at
		FROM reviews
		WHERE product_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.exec(ctx).QueryContext(ctx, query, productID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []review.Review
	for rows.Next() {
		var (
			rev       review.Review
			createdAt string
			updatedAt string
		)
		if err := rows.Scan(&rev.ID, &rev.UserID, &rev.ProductID, &rev.UserName, &rev.Rating, &rev.Comment, &createdAt, &updatedAt); err != nil {
			return nil, 0, err
		}
		rev.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		rev.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		list = append(list, rev)
	}

	return list, total, nil
}

func (r *ReviewRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.exec(ctx).ExecContext(ctx, "DELETE FROM reviews WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}
	return nil
}

func (r *ReviewRepository) GetProductRatingStats(ctx context.Context, productID string) (float64, int, error) {
	query := `SELECT COALESCE(AVG(rating), 0), COUNT(*) FROM reviews WHERE product_id = ?`
	var (
		avg   float64
		count int
	)
	err := r.db.exec(ctx).QueryRowContext(ctx, query, productID).Scan(&avg, &count)
	if err != nil {
		return 0, 0, err
	}
	return avg, count, nil
}
