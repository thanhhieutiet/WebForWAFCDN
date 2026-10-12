package handler

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"

	"labshop/internal/adapter/http/middleware"
	"labshop/internal/usecase/lab"
)

//go:embed openapi.json
var defaultOpenAPIJSON []byte

// LabHandler exposes HTTP endpoints for CDN, WAF, and origin behavior testing.
type LabHandler struct {
	usecase     *lab.Usecase
	openapiJSON []byte
}

// NewLabHandler constructs a LabHandler.
func NewLabHandler(usecase *lab.Usecase, openapiJSON []byte) *LabHandler {
	if len(openapiJSON) == 0 {
		openapiJSON = defaultOpenAPIJSON
	}
	return &LabHandler{
		usecase:     usecase,
		openapiJSON: openapiJSON,
	}
}

// Time handles GET /api/time.
// STRICT REQUIREMENT: DO NOT send Cache-Control, Expires, or ETag to observe CDN default cache.
func (h *LabHandler) Time(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	res := h.usecase.GetTime(reqID)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"generated_at": res.GeneratedAt.Format(time.RFC3339),
		"request_id":   res.RequestID,
	})
}

// Cache handles GET /api/cache/{mode}?v=.
// Modes: public, no-store, etag, vary, set-cookie.
func (h *LabHandler) Cache(w http.ResponseWriter, r *http.Request) {
	mode := chi.URLParam(r, "mode")
	v := r.URL.Query().Get("v")
	reqID := middleware.GetRequestID(r.Context())

	res, err := h.usecase.GetCache(mode, v, reqID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "invalid_mode",
				"message": "Unsupported cache mode",
			},
		})
		return
	}

	switch mode {
	case "public":
		w.Header().Set("Cache-Control", "public, max-age=60")
	case "no-store":
		w.Header().Set("Cache-Control", "no-store")
	case "etag":
		w.Header().Set("ETag", res.ETag)
		ifNoneMatch := r.Header.Get("If-None-Match")
		if ifNoneMatch != "" && (ifNoneMatch == res.ETag || ifNoneMatch == "*") {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	case "vary":
		w.Header().Set("Vary", "Accept-Encoding, Accept-Language")
	case "set-cookie":
		http.SetCookie(w, &http.Cookie{
			Name:     "lab_session",
			Value:    "test",
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"mode":         res.Mode,
		"generated_at": res.GeneratedAt.Format(time.RFC3339),
		"request_id":   res.RequestID,
		"v":            res.Version,
	})
}

// Headers handles GET /api/headers.
// Echoes received headers (masking Authorization and Cookie), client IP, and X-Forwarded-For chain.
func (h *LabHandler) Headers(w http.ResponseWriter, r *http.Request) {
	headers := make(map[string]string)
	for k, v := range r.Header {
		if strings.EqualFold(k, "Authorization") || strings.EqualFold(k, "Cookie") {
			headers[k] = "[REDACTED]"
		} else {
			headers[k] = strings.Join(v, ", ")
		}
	}

	var xff []string
	if rawXFF := r.Header.Get("X-Forwarded-For"); rawXFF != "" {
		parts := strings.Split(rawXFF, ",")
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				xff = append(xff, trimmed)
			}
		}
	}
	if xff == nil {
		xff = []string{}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"headers":         headers,
		"remote_addr":     r.RemoteAddr,
		"x_forwarded_for": xff,
	})
}

// Echo handles GET|POST /api/echo and /api/echo/{path}.
// Echoes query, JSON body, form, header, cookie as JSON.
func (h *LabHandler) Echo(w http.ResponseWriter, r *http.Request) {
	query := make(map[string]string)
	for k, v := range r.URL.Query() {
		query[k] = strings.Join(v, ", ")
	}

	headers := make(map[string]string)
	for k, v := range r.Header {
		headers[k] = strings.Join(v, ", ")
	}

	cookies := make(map[string]string)
	for _, c := range r.Cookies() {
		cookies[c.Name] = c.Value
	}

	bodyBytes, _ := io.ReadAll(io.LimitReader(r.Body, 1024*1024))
	var parsedBody any
	if len(bodyBytes) > 0 {
		_ = json.Unmarshal(bodyBytes, &parsedBody)
	}

	_ = r.ParseForm()
	form := make(map[string]string)
	for k, v := range r.PostForm {
		form[k] = strings.Join(v, ", ")
	}

	subpath := chi.URLParam(r, "path")
	if subpath == "" {
		subpath = chi.URLParam(r, "*")
	}

	resp := map[string]any{
		"path":     r.URL.Path,
		"subpath":  subpath,
		"method":   r.Method,
		"query":    query,
		"headers":  headers,
		"cookies":  cookies,
		"form":     form,
		"body":     parsedBody,
		"raw_body": string(bodyBytes),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// Status handles GET /api/status/{code}.
// Only allows {200, 204, 301, 302, 304, 400, 401, 403, 404, 429, 500, 502, 503, 504}.
func (h *LabHandler) Status(w http.ResponseWriter, r *http.Request) {
	codeStr := chi.URLParam(r, "code")
	code, err := strconv.Atoi(codeStr)
	if err != nil || !h.usecase.IsAllowedStatusCode(code) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "invalid_status_code",
				"message": "Status code not supported",
			},
		})
		return
	}

	if code == http.StatusNoContent || code == http.StatusNotModified {
		w.WriteHeader(code)
		return
	}

	if code == http.StatusMovedPermanently || code == http.StatusFound {
		w.Header().Set("Location", "/api/health")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":   code,
			"location": "/api/health",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  code,
		"message": http.StatusText(code),
	})
}

// Slow handles GET /api/slow?ms=.
// Sleeps context-aware, clamped to MAX_SLOW_MS, max 20 concurrent requests.
func (h *LabHandler) Slow(w http.ResponseWriter, r *http.Request) {
	msStr := r.URL.Query().Get("ms")
	ms := 1000
	if msStr != "" {
		if val, err := strconv.Atoi(msStr); err == nil {
			ms = val
		}
	}

	slept, err := h.usecase.Slow(r.Context(), ms)
	if err != nil {
		if errors.Is(err, lab.ErrSlowBusy) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":    "busy",
					"message": "Server is busy, too many slow requests",
				},
			})
			return
		}
		// Request context cancelled / client aborted
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "ok",
		"slept_ms": slept,
	})
}

// Payload handles GET /api/payload?kb=.
// Returns compressible JSON/text of requested kb (clamped to MAX_PAYLOAD_KB).
func (h *LabHandler) Payload(w http.ResponseWriter, r *http.Request) {
	kbStr := r.URL.Query().Get("kb")
	kb := 1
	if kbStr != "" {
		if val, err := strconv.Atoi(kbStr); err == nil {
			kb = val
		}
	}

	data, err := h.usecase.GeneratePayload(kb)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "internal_error",
				"message": err.Error(),
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"requested_kb": kb,
		"bytes":        len(data),
		"data":         data,
	})
}

// Stats handles GET /api/stats.
// Returns request count per path since startup.
func (h *LabHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats := h.usecase.GetStats()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stats)
}

// ResetStats handles POST /api/stats/reset.
// Checks X-Admin-Token header matching ADMIN_TOKEN from config.
func (h *LabHandler) ResetStats(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Admin-Token")
	if err := h.usecase.ResetStats(token); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "unauthorized",
				"message": "Invalid or missing X-Admin-Token",
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"message": "Stats reset successfully",
	})
}

type uploadResponse struct {
	SHA256      string `json:"sha256"`
	Size        int64  `json:"size"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
}

// Upload handles POST /api/upload.
// Accepts multipart upload, limits size, returns {sha256, size, filename, content_type} without saving to disk.
func (h *LabHandler) Upload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "bad_request",
				"message": "Failed to parse multipart form",
			},
		})
		return
	}

	if r.MultipartForm == nil || len(r.MultipartForm.File) == 0 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "bad_request",
				"message": "No file uploaded in multipart form",
			},
		})
		return
	}

	var fileHeader *multipart.FileHeader
	for _, headers := range r.MultipartForm.File {
		if len(headers) > 0 {
			fileHeader = headers[0]
			break
		}
	}

	file, err := fileHeader.Open()
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "bad_request",
				"message": "Failed to open uploaded file",
			},
		})
		return
	}
	defer file.Close()

	result, err := h.usecase.ProcessUpload(fileHeader.Filename, fileHeader.Header.Get("Content-Type"), file)
	if err != nil {
		if errors.Is(err, lab.ErrUploadTooLarge) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":    "payload_too_large",
					"message": "File exceeds maximum upload limit",
				},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "internal_error",
				"message": err.Error(),
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(uploadResponse{
		SHA256:      result.SHA256,
		Size:        result.Size,
		Filename:    result.Filename,
		ContentType: result.ContentType,
	})
}

// WebSocketEcho handles WS /ws/echo.
func (h *LabHandler) WebSocketEcho(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusInternalError, "internal error")

	c.SetReadLimit(1024 * 1024)

	for {
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		msgType, data, err := c.Read(ctx)
		cancel()
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway ||
				errors.Is(err, context.Canceled) {
				return
			}
			return
		}

		writeCtx, writeCancel := context.WithTimeout(r.Context(), 10*time.Second)
		err = c.Write(writeCtx, msgType, data)
		writeCancel()
		if err != nil {
			return
		}
	}
}

// OpenAPI handles GET /openapi.json.
func (h *LabHandler) OpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(h.openapiJSON)
}

// Robots handles GET /robots.txt.
func (h *LabHandler) Robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("User-agent: *\nDisallow:\n"))
}

// Sitemap handles GET /sitemap.xml.
func (h *LabHandler) Sitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=60")
	w.WriteHeader(http.StatusOK)
	sitemapXML := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>http://localhost:8080/</loc>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>
  <url>
    <loc>http://localhost:8080/products</loc>
    <changefreq>daily</changefreq>
    <priority>0.8</priority>
  </url>
  <url>
    <loc>http://localhost:8080/tutorial</loc>
    <changefreq>weekly</changefreq>
    <priority>0.5</priority>
  </url>
</urlset>`
	_, _ = w.Write([]byte(sitemapXML))
}
