package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"labshop/internal/adapter/counter/memory"
	"labshop/internal/adapter/http/handler"
	"labshop/internal/adapter/http/middleware"
	"labshop/internal/infrastructure/config"
	"labshop/internal/usecase/lab"
	"labshop/internal/usecase/port"
)

// NewRouter constructs a Chi HTTP router, attaches the standard middleware stack,
// registers health, lab, and spec endpoints, and optionally mounts static file servers.
func NewRouter(cfg *config.Config, log *slog.Logger, labHandler *handler.LabHandler, counter port.HitCounter) *chi.Mux {
	r := chi.NewRouter()

	if counter == nil {
		counter = memory.New()
	}

	if labHandler == nil {
		labUC := lab.New(counter, port.RealClock{}, lab.Config{
			AdminToken:   cfg.AdminToken,
			MaxSlowMS:    cfg.MaxSlowMS,
			MaxPayloadKB: cfg.MaxPayloadKB,
			MaxUploadMB:  cfg.MaxUploadMB,
		})
		labHandler = handler.NewLabHandler(labUC, nil)
	}

	// Global middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.Origin(cfg.InstanceID))
	r.Use(middleware.Security)
	r.Use(middleware.HitCounter(counter))
	r.Use(middleware.AccessLog(log))
	r.Use(middleware.Recoverer(log))

	// Health endpoint
	r.Get("/api/health", handler.Health)

	// Lab endpoints
	r.Get("/api/time", labHandler.Time)
	r.Get("/api/cache/{mode}", labHandler.Cache)
	r.Get("/api/headers", labHandler.Headers)
	r.Get("/api/echo", labHandler.Echo)
	r.Post("/api/echo", labHandler.Echo)
	r.Get("/api/echo/*", labHandler.Echo)
	r.Post("/api/echo/*", labHandler.Echo)
	r.Get("/api/status/{code}", labHandler.Status)
	r.Get("/api/slow", labHandler.Slow)
	r.Get("/api/payload", labHandler.Payload)
	r.Get("/api/stats", labHandler.Stats)
	r.Post("/api/stats/reset", labHandler.ResetStats)
	r.Post("/api/upload", labHandler.Upload)

	// WebSocket echo endpoint
	if cfg.EnableWS {
		r.Get("/ws/echo", labHandler.WebSocketEcho)
	}

	// Specification & SEO routes
	r.Get("/openapi.json", labHandler.OpenAPI)
	r.Get("/robots.txt", labHandler.Robots)
	r.Get("/sitemap.xml", labHandler.Sitemap)

	// Static file serving (when SERVE_STATIC=true, e.g. local dev without Nginx)
	if cfg.ServeStatic {
		staticDir := cfg.StaticDir
		if staticDir == "" {
			staticDir = "web/static"
		}
		fs := http.FileServer(http.Dir(staticDir))
		r.Handle("/static/*", http.StripPrefix("/static", fs))
		r.Handle("/contents/*", http.StripPrefix("/contents", fs))
	}

	return r
}
