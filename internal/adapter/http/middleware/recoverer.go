package middleware

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

type panicErrorEnvelope struct {
	Error panicErrorBody `json:"error"`
}

type panicErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// Recoverer intercepts panics, logs the error and stack trace using slog,
// and returns a JSON 500 internal_error response containing the request ID.
func Recoverer(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if log == nil {
				log = slog.Default()
			}

			defer func() {
				if rvr := recover(); rvr != nil {
					if rvr == http.ErrAbortHandler {
						panic(rvr)
					}

					reqID := GetRequestID(r.Context())
					stack := string(debug.Stack())

					log.Error("panic recovered",
						slog.String("request_id", reqID),
						slog.String("error", fmt.Sprint(rvr)),
						slog.String("stack", stack),
						slog.String("path", r.URL.Path),
						slog.String("method", r.Method),
					)

					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusInternalServerError)

					resp := panicErrorEnvelope{
						Error: panicErrorBody{
							Code:      "internal_error",
							Message:   "An unexpected error occurred",
							RequestID: reqID,
						},
					}
					_ = json.NewEncoder(w).Encode(resp)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
