package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/db"
	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/events"
	httpapi "github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/http"
	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/inventory"
)

func main() {
	cfg := loadConfig()
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- DB ---
	pool, err := db.NewPool(ctx, cfg.DatabaseDSN)
	if err != nil {
		logger.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	if cfg.RunMigrations {
		if err := db.RunMigrations(cfg.DatabaseDSN, logger); err != nil {
			logger.Fatalf("db migrate: %v", err)
		}
	}

	repo := inventory.NewPostgresRepository(pool)

	// --- AMQP ---
	conn := events.MustDialRabbit()
	defer conn.Close()

	consumer, cleanupPub, err := events.StartOrderCreatedConsumer(ctx, conn, repo, logger)
	if err != nil {
		logger.Fatalf("start consumer: %v", err)
	}
	defer cleanupPub()

	// --- HTTP ---
	h := httpapi.NewHandler(repo)
	r := httpapi.NewRouter(h)

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		logger.Printf("http listening on %s", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	// --- graceful shutdown ---
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Printf("shutdown signal: %s", sig)
	case err := <-errCh:
		logger.Printf("fatal error: %v", err)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	_ = httpServer.Shutdown(shutdownCtx)
	cancel()

	// best-effort stop consumer loops
	_ = consumer

	logger.Printf("shutdown complete")
}

type config struct {
	HTTPAddr      string
	DatabaseDSN   string
	RunMigrations bool
}

func loadConfig() config {
	return config{
		HTTPAddr:      env("HTTP_ADDR", ":8080"),
		DatabaseDSN:   env("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/inventory?sslmode=disable"),
		RunMigrations: envBool("RUN_MIGRATIONS", true),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "TRUE", "yes", "YES":
		return true
	case "0", "false", "FALSE", "no", "NO":
		return false
	default:
		return fallback
	}
}
