package http

import (
	"context"
	"net/http"
	"time"

	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
)

type OrderHandler struct {
	repo order.Repository
}

func NewOrderHandler(repo order.Repository) *OrderHandler {
	return &OrderHandler{repo: repo}
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("orderId")
	if orderID == "" {
		writeError(w, http.StatusBadRequest, "missing orderId")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	o, err := h.repo.GetByID(ctx, orderID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load order")
		return
	}
	if o == nil {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}

	writeJSON(w, http.StatusOK, o)
}

func (h *OrderHandler) ListOrdersByUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing userId")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orders, err := h.repo.ListByUser(ctx, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load orders")
		return
	}

	writeJSON(w, http.StatusOK, orders)
}
