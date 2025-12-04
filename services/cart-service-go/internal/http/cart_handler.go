package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/andreasstove999/ecommerce-system/cart-service-go/internal/cart"
	"github.com/google/uuid"
)

type CartHandler struct {
	repo           cart.Repository
	eventPublisher CartEventsPublisher
}

type CartEventsPublisher interface {
	PublishCartCheckedOut(ctx context.Context, c *cart.Cart) error
}

func NewCartHandler(repo cart.Repository, eventPublisher CartEventsPublisher) *CartHandler {
	return &CartHandler{repo: repo, eventPublisher: eventPublisher}
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {

	// read {userId} from path, call repo.GetCart, write JSON

	userID := r.PathValue("userId")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing userId")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	c, err := h.repo.GetCart(ctx, userID)
	if err != nil {
		// You can later define your own error types; for now, treat as 500
		writeError(w, http.StatusInternalServerError, "failed to load cart")
		return
	}

	if c == nil {
		writeError(w, http.StatusNotFound, "cart not found")
		return
	}

	writeJSON(w, http.StatusOK, c)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	// parse body (productId, quantity, price), load cart, update, save via repo.UpsertCart
	userID := r.PathValue("userId")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing userId")
		return
	}

	var body struct {
		ProductID string  `json:"productId"`
		Quantity  int     `json:"quantity"`
		Price     float64 `json:"price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// Load existing cart (or nil)
	c, err := h.repo.GetCart(ctx, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load cart")
		return
	}

	// If no cart exists, create new
	if c == nil {
		c = &cart.Cart{
			ID:        uuid.NewString(),
			UserID:    userID,
			Items:     []cart.Item{},
			UpdatedAt: time.Now(),
		}
	}

	// Find existing item or append new
	updated := false
	for i := range c.Items {
		if c.Items[i].ProductID == body.ProductID {
			c.Items[i].Quantity += body.Quantity
			c.Items[i].Price = body.Price
			updated = true
			break
		}
	}
	if !updated {
		c.Items = append(c.Items, cart.Item{
			ProductID: body.ProductID,
			Quantity:  body.Quantity,
			Price:     body.Price,
		})
	}

	// Recalculate total
	total := 0.0
	for _, item := range c.Items {
		total += float64(item.Quantity) * item.Price
	}
	c.Total = total
	c.UpdatedAt = time.Now()

	// Save to DB
	if err := h.repo.UpsertCart(ctx, c); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save cart")
		return
	}

	writeJSON(w, http.StatusOK, c)
}

func (h *CartHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing userId")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	c, err := h.repo.GetCart(ctx, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load cart")
		return
	}
	if c == nil {
		writeError(w, http.StatusNotFound, "cart not found")
		return
	}

	// Publish event (later you add real RabbitMQ)
	if err := h.eventPublisher.PublishCartCheckedOut(ctx, c); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to publish cart checked out event")
		return
	}

	// Clear DB cart
	if err := h.repo.ClearCart(ctx, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to clear cart")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "checkout completed",
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{
		"error": msg,
	})
}
