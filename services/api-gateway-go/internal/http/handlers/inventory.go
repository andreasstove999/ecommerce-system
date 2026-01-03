package handlers

import (
	"net/http"

	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/clients"
)

type InventoryHandler struct{ c *clients.InventoryClient }

func NewInventoryHandler(c *clients.InventoryClient) *InventoryHandler {
	return &InventoryHandler{c: c}
}

func (h *InventoryHandler) Availability(w http.ResponseWriter, r *http.Request) {
	productId := r.PathValue("productId")
	resp, err := h.c.GetAvailability(r.Context(), productId, r.URL.RawQuery, r.Header)
	if err != nil {
		WriteUpstreamError(w, r, http.StatusBadGateway, "inventory-service request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()
	CopyUpstreamResponse(w, resp)
}

func (h *InventoryHandler) Adjust(w http.ResponseWriter, r *http.Request) {
	resp, err := h.c.Adjust(r.Context(), r.URL.RawQuery, r.Body, r.Header)
	if err != nil {
		WriteUpstreamError(w, r, http.StatusBadGateway, "inventory-service request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()
	CopyUpstreamResponse(w, resp)
}
