package middleware

import (
	"context"
	"net/http"
	"strings"

	"labshop/internal/adapter/http/presenter"
	"labshop/internal/domain/errs"
	"labshop/internal/domain/user"
	"labshop/internal/usecase/port"
)

type ctxKeyAuthClaims struct{}

var authClaimsKey = ctxKeyAuthClaims{}

// AuthMiddleware extracts Bearer tokens and attaches claims to the request context if valid.
func AuthMiddleware(issuer port.TokenIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				tokenStr := strings.TrimSpace(authHeader[7:])
				if tokenStr != "" {
					claims, err := issuer.ValidateToken(tokenStr)
					if err == nil && claims != nil {
						ctx := context.WithValue(r.Context(), authClaimsKey, claims)
						r = r.WithContext(ctx)
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetClaims retrieves authenticated claims from the context if present.
func GetClaims(ctx context.Context) *port.Claims {
	if ctx == nil {
		return nil
	}
	claims, _ := ctx.Value(authClaimsKey).(*port.Claims)
	return claims
}

// RequireAuth enforces that a valid user is authenticated.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaims(r.Context())
		if claims == nil {
			presenter.Error(w, r, errs.NewUnauthorized("authentication required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin enforces that the authenticated user is an administrator.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaims(r.Context())
		if claims == nil {
			presenter.Error(w, r, errs.NewUnauthorized("authentication required"))
			return
		}
		if claims.Role != user.RoleAdmin {
			presenter.Error(w, r, errs.NewForbidden("admin access required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
