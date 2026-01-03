package middleware

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/model"
)

func Recover(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Printf("panic: %v", rec)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(model.ErrorResponse{
						Error:         "internal server error",
						CorrelationID: GetCorrelationID(r.Context()),
					})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
