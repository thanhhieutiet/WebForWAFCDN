package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"labshop/internal/domain/user"
	"labshop/internal/usecase/port"
)

type CustomClaims struct {
	UserID string    `json:"uid"`
	Email  string    `json:"email"`
	Role   user.Role `json:"role"`
	jwt.RegisteredClaims
}

type Issuer struct {
	secretKey []byte
	ttl       time.Duration
}

func New(secretKey string, ttl ...time.Duration) *Issuer {
	t := 60 * time.Minute
	if len(ttl) > 0 {
		t = ttl[0]
	}
	return &Issuer{
		secretKey: []byte(secretKey),
		ttl:       t,
	}
}

// IssueToken creates a signed JWT for the user.
func (i *Issuer) IssueToken(u *user.User) (string, error) {
	now := time.Now()
	expiresAt := now.Add(i.ttl)

	claims := CustomClaims{
		UserID: u.ID,
		Email:  u.Email,
		Role:   u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   u.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(i.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenStr, nil
}

// ValidateToken validates the JWT signature and expiration and returns claims.
func (i *Issuer) ValidateToken(tokenStr string) (*port.Claims, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return i.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := parsedToken.Claims.(*CustomClaims)
	if !ok || !parsedToken.Valid {
		return nil, errors.New("invalid token claims")
	}

	var exp time.Time
	if claims.ExpiresAt != nil {
		exp = claims.ExpiresAt.Time
	}

	return &port.Claims{
		UserID:    claims.UserID,
		Role:      claims.Role,
		Email:     claims.Email,
		ExpiresAt: exp,
	}, nil
}
