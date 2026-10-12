package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKeyRequestID struct{}

var requestIDKey = ctxKeyRequestID{}

// RequestID generates a fresh UUID v4 for each request, stores it in context,
// and sets the X-Request-Id response header. Incoming X-Request-Id headers are ignored
// to prevent header spoofing or untrusted tracking IDs.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.NewString()
		w.Header().Set("X-Request-Id", reqID)
		ctx := context.WithValue(r.Context(), requestIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from the context if present.
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}
