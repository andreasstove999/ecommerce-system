package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/db"
	eventserver "github.com/andreasstove999/ecommerce-system/order-service-go/internal/events"
	httpserver "github.com/andreasstove999/ecommerce-system/order-service-go/internal/http"
	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
)

func main() {
	port := getEnv("PORT", "8082")

	logger := log.New(os.Stdout, "[order-service] ", log.LstdFlags|log.Lshortfile)

	// DB
	database := db.MustOpen()
	orderRepo := order.NewRepository(database)

	// RabbitMQ
	rabbitConn := eventserver.MustDialRabbit()
	defer rabbitConn.Close()

	// Context for consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := eventserver.StartCartCheckedOutConsumer(ctx, rabbitConn, orderRepo, logger); err != nil {
		logger.Fatalf("start consumer: %v", err)
	}

	// HTTP
	mux := httpserver.NewRouter(orderRepo)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Printf("order-service listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.Println("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
	cancel()
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
