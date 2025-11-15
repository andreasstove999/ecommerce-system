package http

import (
	"encoding/json"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)

	// Cart-specific endpoints will be added here later.
	// e.g. /api/cart, /api/cart/{userId}/items, /api/cart/{userId}/checkout

	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"status": "ok", "service": "cart-service"}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
