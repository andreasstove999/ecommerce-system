package handlers

import (
	"net/http"

	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/clients"
)

type ShippingHandler struct{ c *clients.ShippingClient }

func NewShippingHandler(c *clients.ShippingClient) *ShippingHandler { return &ShippingHandler{c: c} }

func (h *ShippingHandler) ByOrder(w http.ResponseWriter, r *http.Request) {
	orderId := r.PathValue("orderId")
	resp, err := h.c.ByOrder(r.Context(), orderId, r.URL.RawQuery, r.Header)
	if err != nil {
		WriteUpstreamError(w, r, http.StatusBadGateway, "shipping-service request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()
	CopyUpstreamResponse(w, resp)
}
