package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/andreasstove999/ecommerce-system/cart-service-go/internal/cart"
	"github.com/andreasstove999/ecommerce-system/cart-service-go/internal/db"
	httpserver "github.com/andreasstove999/ecommerce-system/cart-service-go/internal/http"
)

func main() {
	port := getEnv("PORT", "8081") // cart can use 8081

	// Open DB and create repository
	database := db.MustOpen()
	defer database.Close()
	cartRepo := cart.NewRepository(database)

	mux := httpserver.NewRouter(cartRepo)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("cart-service listening on :%s", port)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
