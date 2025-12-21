package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/health", h.Health)

	r.Route("/api/inventory", func(r chi.Router) {
		r.Get("/{productId}", h.GetAvailability)
		r.Post("/adjust", h.AdjustAvailability)
	})

	return r
}
