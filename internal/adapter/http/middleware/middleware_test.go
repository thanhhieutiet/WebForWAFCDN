package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID(t *testing.T) {
	var capturedID string
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// Untrusted incoming request header should be overwritten
	req.Header.Set("X-Request-Id", "malicious-spoofed-id")

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	respID := w.Header().Get("X-Request-Id")
	assert.NotEmpty(t, respID)
	assert.NotEqual(t, "malicious-spoofed-id", respID)
	assert.Equal(t, respID, capturedID)
}

func TestOrigin(t *testing.T) {
	instanceID := "origin-node-01"
	h := Origin(instanceID)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, instanceID, w.Header().Get("X-Origin-Instance"))
	timeHeader := w.Header().Get("X-Origin-Time")
	assert.NotEmpty(t, timeHeader)

	parsedTime, err := time.Parse(RFC3339Milli, timeHeader)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), parsedTime, 2*time.Second)
}

func TestSecurity(t *testing.T) {
	h := Security(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "no-referrer", w.Header().Get("Referrer-Policy"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}

func TestAccessLog(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	h := RequestID(AccessLog(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello world"))
	})))

	req := httptest.NewRequest(http.MethodGet, "/test?param=1", nil)
	req.Header.Set("User-Agent", "TestRunner/1.0")
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	logOutput := buf.String()
	assert.Contains(t, logOutput, `"method":"GET"`)
	assert.Contains(t, logOutput, `"path":"/test"`)
	assert.Contains(t, logOutput, `"query":"param=1"`)
	assert.Contains(t, logOutput, `"status":200`)
	assert.Contains(t, logOutput, `"bytes":11`)
	assert.Contains(t, logOutput, `"xff":"203.0.113.1"`)
	assert.Contains(t, logOutput, `"ua":"TestRunner/1.0"`)
}

func TestRecoverer(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	h := RequestID(Recoverer(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("database crash explosion")
	})))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var body panicErrorEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)

	assert.Equal(t, "internal_error", body.Error.Code)
	assert.Equal(t, "An unexpected error occurred", body.Error.Message)
	assert.NotEmpty(t, body.Error.RequestID)

	assert.Contains(t, buf.String(), "database crash explosion")
	assert.Contains(t, buf.String(), body.Error.RequestID)
}

type testCounter struct {
	counts map[string]int64
}

func (c *testCounter) Increment(path string) {
	c.counts[path]++
}
func (c *testCounter) GetStats() map[string]int64 { return c.counts }
func (c *testCounter) Reset()                     { c.counts = make(map[string]int64) }

func TestHitCounter(t *testing.T) {
	tc := &testCounter{counts: make(map[string]int64)}
	h := HitCounter(tc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/api/time", nil)
	req2 := httptest.NewRequest(http.MethodGet, "/api/time", nil)
	req3 := httptest.NewRequest(http.MethodGet, "/api/health", nil)

	h.ServeHTTP(httptest.NewRecorder(), req1)
	h.ServeHTTP(httptest.NewRecorder(), req2)
	h.ServeHTTP(httptest.NewRecorder(), req3)

	assert.Equal(t, int64(2), tc.counts["/api/time"])
	assert.Equal(t, int64(1), tc.counts["/api/health"])
}

