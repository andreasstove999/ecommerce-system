package handlers

import (
	"net/http"

	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/clients"
	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/middleware"
)

type OrderHandler struct{ c *clients.OrderClient }

func NewOrderHandler(c *clients.OrderClient) *OrderHandler { return &OrderHandler{c: c} }

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderId := r.PathValue("orderId")
	resp, err := h.c.GetOrder(r.Context(), orderId, r.URL.RawQuery, r.Header)
	if err != nil {
		WriteUpstreamError(w, r, http.StatusBadGateway, "order-service request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()
	CopyUpstreamResponse(w, resp)
}

func (h *OrderHandler) ListOrdersMe(w http.ResponseWriter, r *http.Request) {
	userId := middleware.GetUserID(r.Context())
	resp, err := h.c.ListOrdersByUser(r.Context(), userId, r.URL.RawQuery, r.Header)
	if err != nil {
		WriteUpstreamError(w, r, http.StatusBadGateway, "order-service request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()
	CopyUpstreamResponse(w, resp)
}
