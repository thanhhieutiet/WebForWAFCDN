package middleware

import (
	"net/http"

	"labshop/internal/usecase/port"
)

// HitCounter creates a middleware that records the hit count for the request's URL path.
func HitCounter(counter port.HitCounter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if counter != nil {
				counter.Increment(r.URL.Path)
			}
			next.ServeHTTP(w, r)
		})
	}
}
