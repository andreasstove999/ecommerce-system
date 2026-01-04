package main

import (
	"log"
	"net/http"
	"os"
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

	logger.Printf("cart-service listening on :%s", port)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("server error: %v", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
