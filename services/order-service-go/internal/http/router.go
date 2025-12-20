package http

import (
	"encoding/json"
	"net/http"

	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
)

func NewRouter(repo order.Repository) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)

	h := NewOrderHandler(repo)

	mux.HandleFunc("GET /api/orders/{orderId}", h.GetOrder)
	mux.HandleFunc("GET /api/users/{userId}/orders", h.ListOrdersByUser)

	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "order-service",
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
