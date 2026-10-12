package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"labshop/internal/domain/catalog"
	"labshop/internal/domain/money"
	"labshop/internal/usecase/port"
)

type CatalogRepository struct {
	db *DB
}

var _ port.CatalogRepository = (*CatalogRepository)(nil)

func NewCatalogRepository(db *DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

func (r *CatalogRepository) ListCategories(ctx context.Context) ([]catalog.Category, error) {
	rows, err := r.db.exec(ctx).QueryContext(ctx, "SELECT id, name, slug, description, created_at FROM categories ORDER BY name ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	var list []catalog.Category
	for rows.Next() {
		var (
			c         catalog.Category
			createdAt string
		)
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &createdAt); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		list = append(list, c)
	}
	return list, nil
}

func (r *CatalogRepository) GetCategoryByID(ctx context.Context, id string) (*catalog.Category, error) {
	var (
		c         catalog.Category
		createdAt string
	)
	err := r.db.exec(ctx).QueryRowContext(ctx, "SELECT id, name, slug, description, created_at FROM categories WHERE id = ?", id).
		Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("category not found")
	}
	if err != nil {
		return nil, err
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &c, nil
}

func (r *CatalogRepository) CreateCategory(ctx context.Context, c *catalog.Category) error {
	query := `INSERT INTO categories (id, name, slug, description, created_at) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.exec(ctx).ExecContext(ctx, query,
		c.ID, c.Name, c.Slug, c.Description, c.CreatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to insert category: %w", err)
	}
	return nil
}

func (r *CatalogRepository) ListProducts(ctx context.Context, filter port.ProductFilter) ([]catalog.Product, int, error) {
	var (
		whereClauses []string
		args         []any
	)

	// Filter by Category
	if filter.CategoryID != "" {
		whereClauses = append(whereClauses, "category_id = ?")
		args = append(args, filter.CategoryID)
	}

	// Filter by Search Query
	if filter.Query != "" {
		whereClauses = append(whereClauses, "(name LIKE ? OR description LIKE ?)")
		pattern := "%" + filter.Query + "%"
		args = append(args, pattern, pattern)
	}

	// Filter by Price range
	if filter.MinPrice != nil {
		whereClauses = append(whereClauses, "price >= ?")
		args = append(args, filter.MinPrice.Int64())
	}
	if filter.MaxPrice != nil {
		whereClauses = append(whereClauses, "price <= ?")
		args = append(args, filter.MaxPrice.Int64())
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Total count query
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM products %s", whereSQL)
	var total int
	if err := r.db.exec(ctx).QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Sorting
	orderSQL := "ORDER BY created_at DESC"
	switch filter.Sort {
	case "price_asc":
		orderSQL = "ORDER BY price ASC"
	case "price_desc":
		orderSQL = "ORDER BY price DESC"
	case "rating":
		orderSQL = "ORDER BY rating_avg DESC, rating_count DESC"
	case "newest":
		orderSQL = "ORDER BY created_at DESC"
	}

	offset := (filter.Page - 1) * filter.PageSize
	querySQL := fmt.Sprintf(`
		SELECT id, category_id, name, slug, description, price, stock, image_path, status, rating_avg, rating_count, created_at, updated_at
		FROM products
		%s
		%s
		LIMIT ? OFFSET ?
	`, whereSQL, orderSQL)

	queryArgs := append(args, filter.PageSize, offset)
	rows, err := r.db.exec(ctx).QueryContext(ctx, querySQL, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []catalog.Product
	for rows.Next() {
		var (
			p         catalog.Product
			price     int64
			status    string
			createdAt string
			updatedAt string
		)
		err := rows.Scan(
			&p.ID, &p.CategoryID, &p.Name, &p.Slug, &p.Description,
			&price, &p.Stock, &p.ImagePath, &status, &p.RatingAvg, &p.RatingCount,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		p.Price = money.VND(price)
		p.Status = catalog.ProductStatus(status)
		p.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		products = append(products, p)
	}

	return products, total, nil
}

func (r *CatalogRepository) GetProductByID(ctx context.Context, id string) (*catalog.Product, error) {
	query := `
		SELECT id, category_id, name, slug, description, price, stock, image_path, status, rating_avg, rating_count, created_at, updated_at
		FROM products WHERE id = ?
	`
	var (
		p         catalog.Product
		price     int64
		status    string
		createdAt string
		updatedAt string
	)
	err := r.db.exec(ctx).QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.CategoryID, &p.Name, &p.Slug, &p.Description,
		&price, &p.Stock, &p.ImagePath, &status, &p.RatingAvg, &p.RatingCount,
		&createdAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("product not found")
	}
	if err != nil {
		return nil, err
	}
	p.Price = money.VND(price)
	p.Status = catalog.ProductStatus(status)
	p.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &p, nil
}

func (r *CatalogRepository) GetProductBySlug(ctx context.Context, slug string) (*catalog.Product, error) {
	query := `
		SELECT id, category_id, name, slug, description, price, stock, image_path, status, rating_avg, rating_count, created_at, updated_at
		FROM products WHERE slug = ?
	`
	var (
		p         catalog.Product
		price     int64
		status    string
		createdAt string
		updatedAt string
	)
	err := r.db.exec(ctx).QueryRowContext(ctx, query, slug).Scan(
		&p.ID, &p.CategoryID, &p.Name, &p.Slug, &p.Description,
		&price, &p.Stock, &p.ImagePath, &status, &p.RatingAvg, &p.RatingCount,
		&createdAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("product not found")
	}
	if err != nil {
		return nil, err
	}
	p.Price = money.VND(price)
	p.Status = catalog.ProductStatus(status)
	p.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &p, nil
}

func (r *CatalogRepository) CreateProduct(ctx context.Context, p *catalog.Product) error {
	query := `
		INSERT INTO products (id, category_id, name, slug, description, price, stock, image_path, status, rating_avg, rating_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.exec(ctx).ExecContext(ctx, query,
		p.ID, p.CategoryID, p.Name, p.Slug, p.Description,
		p.Price.Int64(), p.Stock, p.ImagePath, string(p.Status),
		p.RatingAvg, p.RatingCount,
		p.CreatedAt.UTC().Format(time.RFC3339),
		p.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to insert product: %w", err)
	}
	return nil
}

func (r *CatalogRepository) UpdateProduct(ctx context.Context, p *catalog.Product) error {
	query := `
		UPDATE products
		SET category_id = ?, name = ?, slug = ?, description = ?, price = ?, stock = ?, image_path = ?, status = ?, rating_avg = ?, rating_count = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.exec(ctx).ExecContext(ctx, query,
		p.CategoryID, p.Name, p.Slug, p.Description,
		p.Price.Int64(), p.Stock, p.ImagePath, string(p.Status),
		p.RatingAvg, p.RatingCount,
		p.UpdatedAt.UTC().Format(time.RFC3339), p.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	return nil
}

func (r *CatalogRepository) DeductStockIfAvailable(ctx context.Context, id string, quantity int) (bool, error) {
	query := `UPDATE products SET stock = stock - ? WHERE id = ? AND stock >= ?`
	res, err := r.db.exec(ctx).ExecContext(ctx, query, quantity, id, quantity)
	if err != nil {
		return false, fmt.Errorf("failed to deduct stock: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (r *CatalogRepository) RestoreStock(ctx context.Context, id string, quantity int) error {
	query := `UPDATE products SET stock = stock + ? WHERE id = ?`
	_, err := r.db.exec(ctx).ExecContext(ctx, query, quantity, id)
	if err != nil {
		return fmt.Errorf("failed to restore stock: %w", err)
	}
	return nil
}

func (r *CatalogRepository) UpdateProductRating(ctx context.Context, id string, ratingAvg float64, ratingCount int) error {
	query := `UPDATE products SET rating_avg = ?, rating_count = ? WHERE id = ?`
	_, err := r.db.exec(ctx).ExecContext(ctx, query, ratingAvg, ratingCount, id)
	if err != nil {
		return fmt.Errorf("failed to update product rating: %w", err)
	}
	return nil
}
