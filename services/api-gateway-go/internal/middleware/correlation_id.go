package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const HeaderCorrelationID = "X-Correlation-Id"

func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cid := r.Header.Get(HeaderCorrelationID)
		if cid == "" {
			cid = uuid.NewString()
		}

		// expose to client + propagate downstream
		w.Header().Set(HeaderCorrelationID, cid)

		ctx := context.WithValue(r.Context(), ctxCorrelationID, cid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetCorrelationID(ctx context.Context) string {
	if v := ctx.Value(ctxCorrelationID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
