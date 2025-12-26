package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/inventory"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	repo inventory.Repository
}

func NewHandler(repo inventory.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (h *Handler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productId")
	item, err := h.repo.Get(r.Context(), productID)
	if err != nil {
		if errors.Is(err, inventory.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

type adjustRequest struct {
	ProductID string `json:"productId"`
	Available int    `json:"available,string"`
}

func (h *Handler) AdjustAvailability(w http.ResponseWriter, r *http.Request) {
	var req adjustRequest
	bodyBytes, _ := io.ReadAll(r.Body)
	log.Println("Request Body:", string(bodyBytes))
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request decoding", http.StatusBadRequest)
		return
	}
	log.Println(req)
	if req.ProductID == "" || req.Available < 0 {
		http.Error(w, "bad request invalid payload", http.StatusBadRequest)
		return
	}

	if err := h.repo.SetAvailable(r.Context(), req.ProductID, req.Available); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
