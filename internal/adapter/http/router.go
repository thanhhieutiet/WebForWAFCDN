package http

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

	"labshop/internal/adapter/http/handler"
	"labshop/internal/adapter/http/middleware"
	"labshop/internal/infrastructure/config"
)

// NewRouter constructs a Chi HTTP router, attaches the standard middleware stack,
// and registers the health check endpoint.
func NewRouter(cfg *config.Config, log *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.Origin(cfg.InstanceID))
	r.Use(middleware.Security)
	r.Use(middleware.AccessLog(log))
	r.Use(middleware.Recoverer(log))

	// Routes
	r.Get("/api/health", handler.Health)

	return r
}
