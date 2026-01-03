package handlers

import (
	"net/http"

	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/clients"
)

type PaymentHandler struct{ c *clients.PaymentClient }

func NewPaymentHandler(c *clients.PaymentClient) *PaymentHandler { return &PaymentHandler{c: c} }

func (h *PaymentHandler) ByOrder(w http.ResponseWriter, r *http.Request) {
	orderId := r.PathValue("orderId")
	resp, err := h.c.ByOrder(r.Context(), orderId, r.URL.RawQuery, r.Header)
	if err != nil {
		WriteUpstreamError(w, r, http.StatusBadGateway, "payment-service request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()
	CopyUpstreamResponse(w, resp)
}
