package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/trewolff/corroboros/internal/api"
	"github.com/trewolff/corroboros/internal/database"
)

func main() {
	// wire your dependencies here (DB, services, config)
	// svc := timestamp.NewService(...)

	db, err := database.SetupDatabase()
	if err != nil {
		log.Fatalf("failed to setup database: %v", err)
	}

	r := api.NewRouter(db)

	srv := &http.Server{
		Addr:    ":" + getEnv("PORT", "8080"),
		Handler: r,
	}

	// run server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Println("server started")

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
