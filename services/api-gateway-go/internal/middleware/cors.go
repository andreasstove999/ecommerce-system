package middleware

import (
	"net/http"
	"strings"
)

func CORS(allowOrigins []string) func(http.Handler) http.Handler {
	allowAll := len(allowOrigins) == 1 && allowOrigins[0] == "*"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Handle preflight
			if r.Method == http.MethodOptions {
				writeCORSHeaders(w, origin, allowOrigins, allowAll)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			writeCORSHeaders(w, origin, allowOrigins, allowAll)
			next.ServeHTTP(w, r)
		})
	}
}

func writeCORSHeaders(w http.ResponseWriter, origin string, allowOrigins []string, allowAll bool) {
	if origin == "" {
		return
	}

	if allowAll {
		// For dev simplicity: reflect origin (works better with browsers than "*")
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else if originAllowed(origin, allowOrigins) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		return
	}

	w.Header().Set("Vary", "Origin")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Correlation-Id, X-User-Id")
}

func originAllowed(origin string, allow []string) bool {
	for _, a := range allow {
		if strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(origin)) {
			return true
		}
	}
	return false
}
