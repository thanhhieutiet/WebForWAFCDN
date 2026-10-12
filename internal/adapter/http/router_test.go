package http

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"labshop/internal/infrastructure/config"
	"labshop/internal/usecase/port"
)

func TestRouter_HealthEndpoint(t *testing.T) {
	cfg := &config.Config{
		InstanceID: "test-node",
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router := NewRouter(cfg, logger, nil, nil)

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

func TestRouter_StaticServing(t *testing.T) {
	tempDir := t.TempDir()
	testFilePath := tempDir + "/hello.txt"
	require.NoError(t, os.WriteFile(testFilePath, []byte("hello static world"), 0644))

	cfg := &config.Config{
		InstanceID:  "test-node",
		ServeStatic: true,
		StaticDir:   tempDir,
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router := NewRouter(cfg, logger, nil, nil)

	// Test /static/hello.txt
	req := httptest.NewRequest(http.MethodGet, "/static/hello.txt", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello static world", w.Body.String())

	// Test /contents/hello.txt (alias for origin path testing)
	reqContents := httptest.NewRequest(http.MethodGet, "/contents/hello.txt", nil)
	wContents := httptest.NewRecorder()
	router.ServeHTTP(wContents, reqContents)
	assert.Equal(t, http.StatusOK, wContents.Code)
	assert.Equal(t, "hello static world", wContents.Body.String())
}

func TestRouter_TutorialScenariosEndpoint(t *testing.T) {
	cfg := &config.Config{
		InstanceID:          "test-node",
		CatalogCacheSeconds: 60,
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router := NewRouter(cfg, logger, nil, nil)

	// Test GET /api/tutorial/scenarios
	req := httptest.NewRequest(http.MethodGet, "/api/tutorial/scenarios", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "public, max-age=60", w.Header().Get("Cache-Control"))
	assert.NotEmpty(t, w.Header().Get("ETag"))

	var list []port.Scenario
	err := json.Unmarshal(w.Body.Bytes(), &list)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 23)

	// Test GET /api/tutorial/scenarios/C1
	reqC1 := httptest.NewRequest(http.MethodGet, "/api/tutorial/scenarios/C1", nil)
	wC1 := httptest.NewRecorder()
	router.ServeHTTP(wC1, reqC1)
	assert.Equal(t, http.StatusOK, wC1.Code)

	var s port.Scenario
	err = json.Unmarshal(wC1.Body.Bytes(), &s)
	require.NoError(t, err)
	assert.Equal(t, "C1", s.ID)

	// Test GET /tutorial/scenarios alias
	reqAlias := httptest.NewRequest(http.MethodGet, "/tutorial/scenarios", nil)
	wAlias := httptest.NewRecorder()
	router.ServeHTTP(wAlias, reqAlias)
	assert.Equal(t, http.StatusOK, wAlias.Code)
}

