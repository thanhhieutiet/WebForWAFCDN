package jwt

import (
	"testing"
	"time"

	"labshop/internal/domain/user"
)

func TestJWTIssuer(t *testing.T) {
	secret := "super-secure-secret-key-32-chars-long!"
	issuer := New(secret, time.Hour)

	u := &user.User{
		ID:    "user-1",
		Email: "user@example.com",
		Role:  user.RoleCustomer,
	}

	tokenStr, err := issuer.IssueToken(u)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	claims, err := issuer.ValidateToken(tokenStr)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != u.ID || claims.Role != u.Role || claims.Email != u.Email {
		t.Fatalf("claims mismatch: %+v", claims)
	}

	// Tampered token
	_, err = issuer.ValidateToken(tokenStr + "bad")
	if err == nil {
		t.Fatalf("expected error for tampered token")
	}

	// Expired token
	shortIssuer := New(secret, -time.Second)
	expiredToken, _ := shortIssuer.IssueToken(u)
	_, err = issuer.ValidateToken(expiredToken)
	if err == nil {
		t.Fatalf("expected error for expired token")
	}
}
