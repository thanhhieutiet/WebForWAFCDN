package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"labshop/internal/adapter/counter/memory"
	"labshop/internal/adapter/http/handler"
	"labshop/internal/adapter/http/middleware"
	"labshop/internal/adapter/scenario/embedded"
	"labshop/internal/infrastructure/config"
	"labshop/internal/usecase/lab"
	"labshop/internal/usecase/port"
	tutorialuc "labshop/internal/usecase/tutorial"
)

// BusinessHandlers bundles handlers for business domain endpoints.
type BusinessHandlers struct {
	Auth        *handler.AuthHandler
	Catalog     *handler.CatalogHandler
	Cart        *handler.CartHandler
	Order       *handler.OrderHandler
	Review      *handler.ReviewHandler
	Support     *handler.SupportHandler
	Admin       *handler.AdminHandler
	Tutorial    *handler.TutorialHandler
	TokenIssuer port.TokenIssuer
}

// NewRouter constructs a Chi HTTP router, attaches the standard middleware stack,
// registers health, lab, spec, and business domain endpoints.
func NewRouter(cfg *config.Config, log *slog.Logger, labHandler *handler.LabHandler, counter port.HitCounter, business ...*BusinessHandlers) *chi.Mux {
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

	var b *BusinessHandlers
	if len(business) > 0 && business[0] != nil {
		b = business[0]
		if b.TokenIssuer != nil {
			r.Use(middleware.AuthMiddleware(b.TokenIssuer))
		}
	}

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

	// Tutorial endpoints
	var tutorialHandler *handler.TutorialHandler
	if b != nil && b.Tutorial != nil {
		tutorialHandler = b.Tutorial
	} else {
		scenarioRepo := embedded.MustNewRepository()
		tutorialUC := tutorialuc.New(scenarioRepo)
		tutorialHandler = handler.NewTutorialHandler(tutorialUC, cfg.CatalogCacheSeconds)
	}
	r.Get("/api/tutorial/scenarios", tutorialHandler.ListScenarios)
	r.Get("/api/tutorial/scenarios/{id}", tutorialHandler.GetScenario)
	r.Get("/tutorial/scenarios", tutorialHandler.ListScenarios)
	r.Get("/tutorial/scenarios/{id}", tutorialHandler.GetScenario)

	// Business domain endpoints
	if b != nil {
		// Auth
		if b.Auth != nil {
			r.Post("/api/auth/register", b.Auth.Register)
			r.Post("/api/auth/login", b.Auth.Login)
			r.Group(func(ar chi.Router) {
				ar.Use(middleware.RequireAuth)
				ar.Get("/api/auth/me", b.Auth.GetMe)
				ar.Patch("/api/auth/me", b.Auth.UpdateMe)
				ar.Post("/api/auth/change-password", b.Auth.ChangePassword)
			})
		}

		// Catalog
		if b.Catalog != nil {
			r.Get("/api/categories", b.Catalog.ListCategories)
			r.Get("/api/products", b.Catalog.ListProducts)
			r.Get("/api/products/{slug}", b.Catalog.GetProductBySlug)
		}

		// Cart
		if b.Cart != nil {
			r.Group(func(cr chi.Router) {
				cr.Use(middleware.RequireAuth)
				cr.Get("/api/cart", b.Cart.GetCart)
				cr.Put("/api/cart/items/{product_id}", b.Cart.UpdateItemQuantity)
				cr.Delete("/api/cart/items/{product_id}", b.Cart.RemoveItem)
				cr.Delete("/api/cart", b.Cart.ClearCart)
				cr.Post("/api/cart/coupon", b.Cart.ApplyCoupon)
				cr.Delete("/api/cart/coupon", b.Cart.RemoveCoupon)
			})
		}

		// Orders
		if b.Order != nil {
			r.Group(func(or chi.Router) {
				or.Use(middleware.RequireAuth)
				or.Post("/api/orders", b.Order.CreateOrder)
				or.Get("/api/orders", b.Order.ListOrders)
				or.Get("/api/orders/{id}", b.Order.GetOrder)
				or.Post("/api/orders/{id}/pay", b.Order.PayOrder)
				or.Post("/api/orders/{id}/cancel", b.Order.CancelOrder)
			})
		}

		// Reviews
		if b.Review != nil {
			r.Get("/api/products/{slug}/reviews", b.Review.ListReviews)
			r.Group(func(rr chi.Router) {
				rr.Use(middleware.RequireAuth)
				rr.Post("/api/products/{slug}/reviews", b.Review.CreateReview)
				rr.Delete("/api/reviews/{id}", b.Review.DeleteReview)
			})
		}

		// Support
		if b.Support != nil {
			r.Group(func(sr chi.Router) {
				sr.Use(middleware.RequireAuth)
				sr.Post("/api/support/tickets", b.Support.CreateTicket)
				sr.Get("/api/support/tickets", b.Support.ListTickets)
				sr.Get("/api/support/tickets/{id}", b.Support.GetTicket)
				sr.Post("/api/support/tickets/{id}/messages", b.Support.AddMessage)
				sr.Post("/api/support/tickets/{id}/attachments", b.Support.AddAttachment)
			})
		}

		// Admin
		if b.Admin != nil {
			r.Group(func(adm chi.Router) {
				adm.Use(middleware.RequireAdmin)
				if b.Catalog != nil {
					adm.Get("/api/admin/products", b.Catalog.ListProducts)
				}
				adm.Post("/api/admin/products", b.Admin.CreateProduct)
				adm.Put("/api/admin/products/{id}", b.Admin.UpdateProduct)
				adm.Patch("/api/admin/products/{id}", b.Admin.UpdateProduct)
				adm.Delete("/api/admin/products/{id}", b.Admin.DeleteProduct)
				adm.Get("/api/admin/orders", b.Admin.ListOrders)
				adm.Patch("/api/admin/orders/{id}/status", b.Admin.UpdateOrderStatus)
				adm.Get("/api/admin/tickets", b.Admin.ListTickets)
				adm.Post("/api/admin/tickets/{id}/reply", b.Admin.ReplyTicket)
				adm.Get("/api/admin/users", b.Admin.ListUsers)
				adm.Get("/api/admin/stats", b.Admin.GetStats)
			})
		}
	}

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
