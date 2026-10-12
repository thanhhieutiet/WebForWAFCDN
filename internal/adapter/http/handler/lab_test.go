package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"labshop/internal/adapter/counter/memory"
	"labshop/internal/adapter/http/middleware"
	"labshop/internal/usecase/lab"
)

type fixedTestClock struct {
	now time.Time
}

func (f fixedTestClock) Now() time.Time {
	return f.now
}

func setupTestLab() (*chi.Mux, *LabHandler, *memory.MemoryHitCounter, string) {
	adminToken := "test-admin-secret-token"
	counter := memory.New()
	fixedNow := time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC)

	uc := lab.New(counter, fixedTestClock{now: fixedNow}, lab.Config{
		AdminToken:   adminToken,
		MaxSlowMS:    5000,
		MaxPayloadKB: 512,
		MaxUploadMB:  2,
	})

	handler := NewLabHandler(uc, defaultOpenAPIJSON)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.HitCounter(counter))

	r.Get("/api/time", handler.Time)
	r.Get("/api/cache/{mode}", handler.Cache)
	r.Get("/api/headers", handler.Headers)
	r.Get("/api/echo", handler.Echo)
	r.Post("/api/echo", handler.Echo)
	r.Get("/api/echo/*", handler.Echo)
	r.Post("/api/echo/*", handler.Echo)
	r.Get("/api/status/{code}", handler.Status)
	r.Get("/api/slow", handler.Slow)
	r.Get("/api/payload", handler.Payload)
	r.Get("/api/stats", handler.Stats)
	r.Post("/api/stats/reset", handler.ResetStats)
	r.Post("/api/upload", handler.Upload)
	r.Get("/ws/echo", handler.WebSocketEcho)
	r.Get("/openapi.json", handler.OpenAPI)
	r.Get("/robots.txt", handler.Robots)
	r.Get("/sitemap.xml", handler.Sitemap)

	return r, handler, counter, adminToken
}

func TestTimeEndpoint(t *testing.T) {
	r, _, _, _ := setupTestLab()

	req := httptest.NewRequest(http.MethodGet, "/api/time", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// CRITICAL: DO NOT send Cache-Control, Expires, or ETag
	assert.Empty(t, w.Header().Get("Cache-Control"), "must not send Cache-Control")
	assert.Empty(t, w.Header().Get("Expires"), "must not send Expires")
	assert.Empty(t, w.Header().Get("ETag"), "must not send ETag")

	var body map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.NotEmpty(t, body["generated_at"])
	assert.NotEmpty(t, body["request_id"])
}

func TestCacheEndpoint(t *testing.T) {
	r, _, _, _ := setupTestLab()

	// 1. public
	req := httptest.NewRequest(http.MethodGet, "/api/cache/public?v=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "public, max-age=60", w.Header().Get("Cache-Control"))
	var bodyPub map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &bodyPub))
	assert.Equal(t, "public", bodyPub["mode"])
	assert.Equal(t, "1", bodyPub["v"])

	// 2. no-store
	req = httptest.NewRequest(http.MethodGet, "/api/cache/no-store", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))

	// 3. etag & 304 handling
	req = httptest.NewRequest(http.MethodGet, "/api/cache/etag?v=2", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	etag := w.Header().Get("ETag")
	assert.NotEmpty(t, etag)

	// Send matching If-None-Match -> should return 304
	req304 := httptest.NewRequest(http.MethodGet, "/api/cache/etag?v=2", nil)
	req304.Header.Set("If-None-Match", etag)
	w304 := httptest.NewRecorder()
	r.ServeHTTP(w304, req304)
	assert.Equal(t, http.StatusNotModified, w304.Code)
	assert.Empty(t, w304.Body.String(), "304 response must have no body")

	// 4. vary
	req = httptest.NewRequest(http.MethodGet, "/api/cache/vary", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Accept-Encoding, Accept-Language", w.Header().Get("Vary"))

	// 5. set-cookie
	req = httptest.NewRequest(http.MethodGet, "/api/cache/set-cookie", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	cookieHeader := w.Header().Get("Set-Cookie")
	assert.Contains(t, cookieHeader, "lab_session=test")
	assert.Contains(t, cookieHeader, "HttpOnly")

	// 6. invalid mode
	req = httptest.NewRequest(http.MethodGet, "/api/cache/unknown-mode", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHeadersEndpoint(t *testing.T) {
	r, _, _, _ := setupTestLab()

	req := httptest.NewRequest(http.MethodGet, "/api/headers", nil)
	req.Header.Set("Authorization", "Bearer sensitive-jwt-token")
	req.Header.Set("Cookie", "session=super-secret-cookie")
	req.Header.Set("X-Custom-Header", "HelloWAF")
	req.Header.Set("X-Forwarded-For", "203.0.113.195, 198.51.100.1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Headers       map[string]string `json:"headers"`
		RemoteAddr    string            `json:"remote_addr"`
		XForwardedFor []string          `json:"x_forwarded_for"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)

	assert.Equal(t, "[REDACTED]", body.Headers["Authorization"])
	assert.Equal(t, "[REDACTED]", body.Headers["Cookie"])
	assert.Equal(t, "HelloWAF", body.Headers["X-Custom-Header"])
	assert.Equal(t, []string{"203.0.113.195", "198.51.100.1"}, body.XForwardedFor)
}

func TestEchoEndpoint(t *testing.T) {
	r, _, _, _ := setupTestLab()

	// GET with query
	req := httptest.NewRequest(http.MethodGet, "/api/echo?name=test&category=tech", nil)
	req.Header.Set("X-Trace", "trace-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var getBody map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &getBody))
	assert.Equal(t, "GET", getBody["method"])
	queryMap := getBody["query"].(map[string]any)
	assert.Equal(t, "test", queryMap["name"])

	// POST with JSON body and subpath
	postData := `{"message":"hello world","items":[1,2,3]}`
	reqPost := httptest.NewRequest(http.MethodPost, "/api/echo/v1/nested/path", strings.NewReader(postData))
	reqPost.Header.Set("Content-Type", "application/json")
	wPost := httptest.NewRecorder()
	r.ServeHTTP(wPost, reqPost)

	assert.Equal(t, http.StatusOK, wPost.Code)
	var postResp map[string]any
	require.NoError(t, json.Unmarshal(wPost.Body.Bytes(), &postResp))
	assert.Equal(t, "POST", postResp["method"])
	assert.Equal(t, "/api/echo/v1/nested/path", postResp["path"])
	bodyMap := postResp["body"].(map[string]any)
	assert.Equal(t, "hello world", bodyMap["message"])
}

func TestStatusEndpoint(t *testing.T) {
	r, _, _, _ := setupTestLab()

	// 200
	req := httptest.NewRequest(http.MethodGet, "/api/status/200", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 204 No Content
	req = httptest.NewRequest(http.MethodGet, "/api/status/204", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())

	// 301 Moved Permanently with Location: /api/health
	req = httptest.NewRequest(http.MethodGet, "/api/status/301", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "/api/health", w.Header().Get("Location"))

	// 302 Found with Location: /api/health
	req = httptest.NewRequest(http.MethodGet, "/api/status/302", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/api/health", w.Header().Get("Location"))

	// 304 Not Modified
	req = httptest.NewRequest(http.MethodGet, "/api/status/304", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotModified, w.Code)
	assert.Empty(t, w.Body.String())

	// 400, 401, 403, 404, 429, 500, 502, 503, 504
	for _, code := range []int{400, 401, 403, 404, 429, 500, 502, 503, 504} {
		req = httptest.NewRequest(http.MethodGet, "/api/status/"+chi.URLParamFromCtx(req.Context(), "code")+strings.TrimPrefix(httptest.NewRequest(http.MethodGet, "/api/status/"+string(rune(code)), nil).URL.Path, "/api/status/"), nil)
		reqTest := httptest.NewRequest(http.MethodGet, "/api/status/"+chi.URLParamFromCtx(req.Context(), "code"), nil)
		_ = reqTest
	}

	// Invalid status code (e.g. 201, 418, 999, invalid) -> 400
	req = httptest.NewRequest(http.MethodGet, "/api/status/418", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/status/abc", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSlowEndpoint(t *testing.T) {
	r, _, _, _ := setupTestLab()

	req := httptest.NewRequest(http.MethodGet, "/api/slow?ms=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "ok", body["status"])
	assert.Equal(t, float64(10), body["slept_ms"])
}

func TestPayloadEndpoint(t *testing.T) {
	r, _, _, _ := setupTestLab()

	req := httptest.NewRequest(http.MethodGet, "/api/payload?kb=4", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, float64(4), body["requested_kb"])
	assert.Equal(t, float64(4*1024), body["bytes"])
	assert.Equal(t, 4*1024, len(body["data"].(string)))
}

func TestStatsAndResetEndpoints(t *testing.T) {
	r, _, counter, adminToken := setupTestLab()

	// Perform requests to increment hits
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/time", nil))
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/time", nil))

	// Get stats
	reqStats := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	wStats := httptest.NewRecorder()
	r.ServeHTTP(wStats, reqStats)

	assert.Equal(t, http.StatusOK, wStats.Code)
	var stats map[string]int64
	require.NoError(t, json.Unmarshal(wStats.Body.Bytes(), &stats))
	assert.Equal(t, int64(2), stats["/api/time"])

	// Reset without token -> 401
	reqResetFail := httptest.NewRequest(http.MethodPost, "/api/stats/reset", nil)
	wResetFail := httptest.NewRecorder()
	r.ServeHTTP(wResetFail, reqResetFail)
	assert.Equal(t, http.StatusUnauthorized, wResetFail.Code)
	assert.NotEmpty(t, counter.GetStats())

	// Reset with valid token -> 200
	reqResetOk := httptest.NewRequest(http.MethodPost, "/api/stats/reset", nil)
	reqResetOk.Header.Set("X-Admin-Token", adminToken)
	wResetOk := httptest.NewRecorder()
	r.ServeHTTP(wResetOk, reqResetOk)
	assert.Equal(t, http.StatusOK, wResetOk.Code)
	assert.Empty(t, counter.GetStats())
}

func TestUploadEndpoint(t *testing.T) {
	r, _, _, _ := setupTestLab()

	// 1. Success upload
	var b bytes.Buffer
	wMulti := multipart.NewWriter(&b)
	part, err := wMulti.CreateFormFile("file", "test-document.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte("LabShop test upload content for SHA256"))
	require.NoError(t, err)
	require.NoError(t, wMulti.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &b)
	req.Header.Set("Content-Type", wMulti.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var uploadRes map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &uploadRes))
	assert.Equal(t, "test-document.txt", uploadRes["filename"])
	assert.NotEmpty(t, uploadRes["sha256"])
	assert.Equal(t, float64(len("LabShop test upload content for SHA256")), uploadRes["size"])

	// 2. Empty form -> 400
	reqEmpty := httptest.NewRequest(http.MethodPost, "/api/upload", strings.NewReader("not a multipart form"))
	reqEmpty.Header.Set("Content-Type", "text/plain")
	wEmpty := httptest.NewRecorder()
	r.ServeHTTP(wEmpty, reqEmpty)
	assert.Equal(t, http.StatusBadRequest, wEmpty.Code)
}

func TestStaticSpecAndRobotsEndpoints(t *testing.T) {
	r, _, _, _ := setupTestLab()

	// /openapi.json
	reqOpenAPI := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	wOpenAPI := httptest.NewRecorder()
	r.ServeHTTP(wOpenAPI, reqOpenAPI)
	assert.Equal(t, http.StatusOK, wOpenAPI.Code)
	assert.Equal(t, "application/json; charset=utf-8", wOpenAPI.Header().Get("Content-Type"))
	var openapiMap map[string]any
	require.NoError(t, json.Unmarshal(wOpenAPI.Body.Bytes(), &openapiMap))
	assert.Equal(t, "3.0.3", openapiMap["openapi"])

	// /robots.txt
	reqRobots := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	wRobots := httptest.NewRecorder()
	r.ServeHTTP(wRobots, reqRobots)
	assert.Equal(t, http.StatusOK, wRobots.Code)
	assert.Contains(t, wRobots.Body.String(), "User-agent: *")
	assert.Contains(t, wRobots.Body.String(), "Disallow:")

	// /sitemap.xml
	reqSitemap := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	wSitemap := httptest.NewRecorder()
	r.ServeHTTP(wSitemap, reqSitemap)
	assert.Equal(t, http.StatusOK, wSitemap.Code)
	assert.Equal(t, "application/xml; charset=utf-8", wSitemap.Header().Get("Content-Type"))
	assert.Equal(t, "public, max-age=60", wSitemap.Header().Get("Cache-Control"))
	assert.Contains(t, wSitemap.Body.String(), "<urlset")
}

func TestWebSocketEcho(t *testing.T) {
	r, _, _, _ := setupTestLab()
	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/echo"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "done")

	msg := "Hello, LabShop WebSocket!"
	err = conn.Write(ctx, websocket.MessageText, []byte(msg))
	require.NoError(t, err)

	typ, received, err := conn.Read(ctx)
	require.NoError(t, err)
	assert.Equal(t, websocket.MessageText, typ)
	assert.Equal(t, msg, string(received))
}
