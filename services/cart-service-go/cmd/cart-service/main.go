package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andreasstove999/ecommerce-system/cart-service-go/internal/cart"
	"github.com/andreasstove999/ecommerce-system/cart-service-go/internal/db"
	"github.com/andreasstove999/ecommerce-system/cart-service-go/internal/events"
	httpserver "github.com/andreasstove999/ecommerce-system/cart-service-go/internal/http"
)

func main() {
	port := getEnv("PORT", "8081") // cart can use 8081

	logger := log.New(os.Stdout, "[cart-service] ", log.LstdFlags|log.Lshortfile)

	dsn := db.GetDSN()
	if err := db.RunMigrations(dsn, logger); err != nil {
		logger.Fatalf("run migrations: %v", err)
	}

	// Open DB and create repository
	database := db.MustOpen()
	defer database.Close()
	cartRepo := cart.NewRepository(database)

	rabbitConn := events.MustDialRabbit()
	defer rabbitConn.Close()

	sequenceRepo := events.NewSequenceRepository(database)
	cartPublisher, err := events.NewRabbitCartEventsPublisher(rabbitConn, sequenceRepo)
	if err != nil {
		logger.Fatalf("failed to create cart publisher: %v", err)
	}

	mux := httpserver.NewRouter(cartRepo, cartPublisher)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Printf("cart-service listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Printf("shutdown signal received")
	case err := <-errCh:
		logger.Fatalf("server error: %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Printf("graceful shutdown error: %v", err)
	}
	if err := cartPublisher.Close(); err != nil {
		logger.Printf("publisher close error: %v", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
