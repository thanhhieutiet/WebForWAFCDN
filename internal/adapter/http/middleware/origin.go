package middleware

import (
	"net/http"
	"time"
)

// RFC3339Milli defines RFC3339 format with millisecond precision.
const RFC3339Milli = "2006-01-02T15:04:05.000Z07:00"

// Origin returns a middleware that sets X-Origin-Instance and X-Origin-Time (RFC3339 ms) headers.
func Origin(instanceID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Origin-Instance", instanceID)
			w.Header().Set("X-Origin-Time", time.Now().UTC().Format(RFC3339Milli))
			next.ServeHTTP(w, r)
		})
	}
}
