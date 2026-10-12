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

	httpadapter "labshop/internal/adapter/http"
	"labshop/internal/infrastructure/config"
	"labshop/internal/infrastructure/logger"
)

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
	)

	router := httpadapter.NewRouter(cfg, log)

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
