package user

import (
	"testing"
	"time"
)

func TestUserValidation(t *testing.T) {
	u := &User{
		ID:        "u-1",
		Email:     "user@example.com",
		Role:      RoleCustomer,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := u.Validate(); err != nil {
		t.Fatalf("expected valid user, got error: %v", err)
	}

	if u.IsAdmin() {
		t.Fatalf("expected IsAdmin to be false")
	}

	u.Role = RoleAdmin
	if !u.IsAdmin() {
		t.Fatalf("expected IsAdmin to be true")
	}

	u.Email = "invalid-email"
	if err := u.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid email")
	}

	u.Email = "test@example.com"
	u.Role = "superman"
	if err := u.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid role")
	}
}
