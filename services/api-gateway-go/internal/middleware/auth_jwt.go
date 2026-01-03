package middleware

import "net/http"

// AuthJWT is intentionally a no-op placeholder for Demo purposes.
// Can be made more secure later if needed.

func AuthJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
