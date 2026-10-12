package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"labshop/internal/domain/user"
	"labshop/internal/usecase/port"
)

type UserRepository struct {
	db *DB
}

var _ port.UserRepository = (*UserRepository)(nil)

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, name, phone, address, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.exec(ctx).ExecContext(ctx, query,
		u.ID, u.Email, u.PasswordHash, u.Name, u.Phone, u.Address, string(u.Role),
		u.CreatedAt.UTC().Format(time.RFC3339),
		u.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, name, phone, address, role, created_at, updated_at
		FROM users WHERE id = ?
	`
	var (
		u         user.User
		roleStr   string
		createdAt string
		updatedAt string
	)
	err := r.db.exec(ctx).QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.Address, &roleStr, &createdAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	u.Role = user.Role(roleStr)
	u.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	u.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, name, phone, address, role, created_at, updated_at
		FROM users WHERE LOWER(email) = LOWER(?)
	`
	var (
		u         user.User
		roleStr   string
		createdAt string
		updatedAt string
	)
	err := r.db.exec(ctx).QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.Address, &roleStr, &createdAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	u.Role = user.Role(roleStr)
	u.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	u.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &u, nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	query := `
		UPDATE users
		SET email = ?, password_hash = ?, name = ?, phone = ?, address = ?, role = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.exec(ctx).ExecContext(ctx, query,
		u.Email, u.PasswordHash, u.Name, u.Phone, u.Address, string(u.Role),
		u.UpdatedAt.UTC().Format(time.RFC3339), u.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]user.User, int, error) {
	var total int
	err := r.db.exec(ctx).QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	query := `
		SELECT id, email, password_hash, name, phone, address, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.exec(ctx).QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []user.User
	for rows.Next() {
		var (
			u         user.User
			roleStr   string
			createdAt string
			updatedAt string
		)
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.Address, &roleStr, &createdAt, &updatedAt); err != nil {
			return nil, 0, err
		}
		u.Role = user.Role(roleStr)
		u.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		u.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		users = append(users, u)
	}

	return users, total, nil
}
