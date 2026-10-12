package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"

	"labshop/internal/adapter/counter/memory"
	httpadapter "labshop/internal/adapter/http"
	"labshop/internal/adapter/http/handler"
	mockpayment "labshop/internal/adapter/payment/mock"
	"labshop/internal/adapter/repository/sqlite"
	"labshop/internal/adapter/scenario/embedded"
	"labshop/internal/adapter/security/argon2"
	"labshop/internal/adapter/security/jwt"
	"labshop/internal/infrastructure/config"
	infraDB "labshop/internal/infrastructure/db"
	"labshop/internal/infrastructure/logger"
	adminuc "labshop/internal/usecase/admin"
	authuc "labshop/internal/usecase/auth"
	cartuc "labshop/internal/usecase/cart"
	cataloguc "labshop/internal/usecase/catalog"
	"labshop/internal/usecase/lab"
	orderuc "labshop/internal/usecase/order"
	"labshop/internal/usecase/port"
	reviewuc "labshop/internal/usecase/review"
	supportuc "labshop/internal/usecase/support"
	tutorialuc "labshop/internal/usecase/tutorial"
)

type uuidGenerator struct{}

func (uuidGenerator) NewID() string {
	return uuid.NewString()
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel)
	log.Info("initializing labshop server",
		slog.String("app_env", cfg.AppEnv),
		slog.String("http_addr", cfg.HTTPAddr),
		slog.String("instance_id", cfg.InstanceID),
		slog.Bool("serve_static", cfg.ServeStatic),
		slog.String("db_path", cfg.DBPath),
		slog.Bool("seed", cfg.Seed),
	)

	// Database Initialization & Migrations
	dbInstance, err := infraDB.Init(cfg.DBPath)
	if err != nil {
		log.Error("failed to initialize sqlite database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbInstance.Close()

	// Infrastructure & Security Adapters
	idGen := uuidGenerator{}
	hasher := argon2.New()
	jwtIssuer := jwt.New(cfg.JWTSecret)
	paymentGw := mockpayment.New(cfg.MockPaymentFailRate)
	txManager := sqlite.NewTxManager(dbInstance)

	// SQLite Repositories
	userRepo := sqlite.NewUserRepository(dbInstance)
	catalogRepo := sqlite.NewCatalogRepository(dbInstance)
	cartRepo := sqlite.NewCartRepository(dbInstance)
	couponRepo := sqlite.NewCouponRepository(dbInstance)
	orderRepo := sqlite.NewOrderRepository(dbInstance)
	reviewRepo := sqlite.NewReviewRepository(dbInstance)
	supportRepo := sqlite.NewSupportRepository(dbInstance)
	adminRepo := sqlite.NewAdminRepository(dbInstance)

	// Seeder
	if cfg.Seed {
		log.Info("checking database seed requirements")
		if err := infraDB.Seed(
			context.Background(),
			dbInstance,
			hasher,
			cfg.SeedAdminEmail,
			cfg.SeedAdminPassword,
			cfg.SeedCustomerEmail,
			cfg.SeedCustomerPassword,
		); err != nil {
			log.Error("failed to seed database", slog.String("error", err.Error()))
			os.Exit(1)
		}
		log.Info("seeding completed or already present")
	}

	// Business Usecases
	authUC := authuc.New(userRepo, hasher, jwtIssuer, idGen)
	catalogUC := cataloguc.New(catalogRepo)
	cartUC := cartuc.New(cartRepo, catalogRepo, couponRepo)
	orderUC := orderuc.New(orderRepo, cartRepo, catalogRepo, couponRepo, paymentGw, txManager, idGen)
	reviewUC := reviewuc.New(reviewRepo, orderRepo, catalogRepo, idGen)
	supportUC := supportuc.New(supportRepo, idGen)
	adminUC := adminuc.New(catalogRepo, orderRepo, supportRepo, userRepo, adminRepo, txManager, idGen)
	scenarioRepo := embedded.MustNewRepository()
	tutorialUC := tutorialuc.New(scenarioRepo)

	// HTTP Handlers
	authHandler := handler.NewAuthHandler(authUC)
	catalogHandler := handler.NewCatalogHandler(catalogUC)
	cartHandler := handler.NewCartHandler(cartUC)
	orderHandler := handler.NewOrderHandler(orderUC)
	reviewHandler := handler.NewReviewHandler(reviewUC)
	supportHandler := handler.NewSupportHandler(supportUC)
	adminHandler := handler.NewAdminHandler(adminUC)
	tutorialHandler := handler.NewTutorialHandler(tutorialUC, cfg.CatalogCacheSeconds)

	businessHandlers := &httpadapter.BusinessHandlers{
		Auth:        authHandler,
		Catalog:     catalogHandler,
		Cart:        cartHandler,
		Order:       orderHandler,
		Review:      reviewHandler,
		Support:     supportHandler,
		Admin:       adminHandler,
		Tutorial:    tutorialHandler,
		TokenIssuer: jwtIssuer,
	}

	// Phase 1 Lab Usecase & Handlers
	hitCounter := memory.New()
	labUsecase := lab.New(hitCounter, port.RealClock{}, lab.Config{
		AdminToken:   cfg.AdminToken,
		MaxSlowMS:    cfg.MaxSlowMS,
		MaxPayloadKB: cfg.MaxPayloadKB,
		MaxUploadMB:  cfg.MaxUploadMB,
	})
	labHandler := handler.NewLabHandler(labUsecase, nil)

	// HTTP Router
	router := httpadapter.NewRouter(cfg, log, labHandler, hitCounter, businessHandlers)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Info("server listening", slog.String("addr", cfg.HTTPAddr))
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server encountered fatal error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	case sig := <-shutdown:
		log.Info("shutdown signal received", slog.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("graceful shutdown failed, forcing close", slog.String("error", err.Error()))
			_ = srv.Close()
			os.Exit(1)
		}
		log.Info("server exited cleanly")
	}
}
