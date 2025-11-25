package http

import (
	"encoding/json"
	"net/http"

	"github.com/andreasstove999/ecommerce-system/cart-service-go/internal/cart"
)

func NewRouter(cartRepo cart.Repository) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	// Wiring for cart
	cartHandler := NewCartHandler(cartRepo)

	mux.HandleFunc("POST /api/cart/{userId}/items", cartHandler.AddItem)     // add/update item
	mux.HandleFunc("GET /api/cart/{userId}", cartHandler.GetCart)            // fetch cart
	mux.HandleFunc("POST /api/cart/{userId}/checkout", cartHandler.Checkout) // publish event
	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"status": "ok", "service": "cart-service"}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
