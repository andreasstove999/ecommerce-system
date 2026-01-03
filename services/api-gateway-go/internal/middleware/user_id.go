package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/model"
)

const HeaderUserID = "X-User-Id"

type ctxKey string

const (
	ctxCorrelationID ctxKey = "correlation_id" // keep your existing one
	ctxUserID        ctxKey = "user_id"
)

func UserIDKey() any { return ctxUserID }

// RequireUserIDForMeRoutes enforces X-User-Id on all /me/* routes and stores it in context.
func RequireUserIDForMeRoutes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/me" || strings.HasPrefix(path, "/me/") {
			uid := strings.TrimSpace(r.Header.Get(HeaderUserID))
			if uid == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(model.ErrorResponse{
					Error:         "missing required header: X-User-Id",
					CorrelationID: GetCorrelationID(r.Context()),
				})
				return
			}
			ctx := context.WithValue(r.Context(), ctxUserID, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func GetUserID(ctx context.Context) string {
	if v := ctx.Value(ctxUserID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
