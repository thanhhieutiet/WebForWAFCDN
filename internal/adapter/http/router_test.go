package http

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"labshop/internal/infrastructure/config"
)

func TestRouter_HealthEndpoint(t *testing.T) {
	cfg := &config.Config{
		InstanceID: "test-node",
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router := NewRouter(cfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-Id"))
	assert.Equal(t, "test-node", w.Header().Get("X-Origin-Instance"))
	assert.NotEmpty(t, w.Header().Get("X-Origin-Time"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "no-referrer", w.Header().Get("Referrer-Policy"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "ok", body["status"])
}
