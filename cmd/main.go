package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/trewolff/corroboros/internal/api"
	"github.com/trewolff/corroboros/internal/config"
	"github.com/trewolff/corroboros/internal/database"

	"github.com/caarlos0/env/v11"
)

func main() {
	var cfg config.Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("failed to parse env: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	db, err := database.SetupDatabase()
	if err != nil {
		logger.Error("failed to setup database", "error", err)
		os.Exit(1)
	}

	handlerDependencies := &api.HandlerDependencies{
		DB:            db,
		Logger:        logger,
		MaxIntakeSize: cfg.MaxUploadSize,
	}
	r := api.NewRouter(handlerDependencies)

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	// run server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen", "error", err)
		}
	}()
	logger.Info("server started")

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}
}
