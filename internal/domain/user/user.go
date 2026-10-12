package user

import (
	"strings"
	"time"

	"labshop/internal/domain/errs"
)

type Role string

const (
	RoleCustomer Role = "customer"
	RoleAdmin    Role = "admin"
)

// User represents a user account in the system.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	Name         string
	Phone        string
	Address      string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Validate checks business rules for a user entity.
func (u *User) Validate() error {
	var violations []errs.FieldViolation

	email := strings.TrimSpace(u.Email)
	if email == "" {
		violations = append(violations, errs.FieldViolation{Field: "email", Issue: "email is required"})
	} else if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		violations = append(violations, errs.FieldViolation{Field: "email", Issue: "invalid email format"})
	}

	if u.Role != RoleCustomer && u.Role != RoleAdmin {
		violations = append(violations, errs.FieldViolation{Field: "role", Issue: "invalid user role"})
	}

	if len(violations) > 0 {
		return errs.NewValidation("user validation failed", violations...)
	}
	return nil
}

// IsAdmin returns true if the user has the admin role.
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
