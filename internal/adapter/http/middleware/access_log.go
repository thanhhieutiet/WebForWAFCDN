package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseRecorder) WriteHeader(code int) {
	if rw.statusCode == 0 {
		rw.statusCode = code
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseRecorder) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

func (rw *responseRecorder) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// AccessLog records incoming HTTP requests in structured slog JSON format.
// Log fields: ts, level, request_id, method, path, query (truncated to 512 chars),
// status, bytes, duration_ms, remote_addr, xff, ua, host, proto.
// Headers like Authorization and Cookie are intentionally omitted for privacy.
func AccessLog(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if log == nil {
				log = slog.Default()
			}

			start := time.Now()
			rec := &responseRecorder{ResponseWriter: w}

			next.ServeHTTP(rec, r)

			duration := time.Since(start)
			durationMs := float64(duration.Microseconds()) / 1000.0

			status := rec.statusCode
			if status == 0 {
				status = http.StatusOK
			}

			query := r.URL.RawQuery
			if len(query) > 512 {
				query = query[:512]
			}

			reqID := GetRequestID(r.Context())
			xff := r.Header.Get("X-Forwarded-For")
			ua := r.UserAgent()

			var logLevel slog.Level
			switch {
			case status >= 500:
				logLevel = slog.LevelError
			case status >= 400:
				logLevel = slog.LevelWarn
			default:
				logLevel = slog.LevelInfo
			}

			log.Log(r.Context(), logLevel, "http access",
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("query", query),
				slog.Int("status", status),
				slog.Int64("bytes", rec.bytesWritten),
				slog.Float64("duration_ms", durationMs),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("xff", xff),
				slog.String("ua", ua),
				slog.String("host", r.Host),
				slog.String("proto", r.Proto),
			)
		})
	}
}
