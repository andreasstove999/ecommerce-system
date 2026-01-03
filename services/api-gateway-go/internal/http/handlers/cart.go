package handlers

import (
	"net/http"

	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/clients"
	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/middleware"
)

type CartHandler struct{ c *clients.CartClient }

func NewCartHandler(c *clients.CartClient) *CartHandler { return &CartHandler{c: c} }

func (h *CartHandler) AddItemMe(w http.ResponseWriter, r *http.Request) {
	userId := middleware.GetUserID(r.Context())
	resp, err := h.c.AddItem(r.Context(), userId, r.URL.RawQuery, r.Body, r.Header)
	if err != nil {
		WriteUpstreamError(w, r, http.StatusBadGateway, "cart-service request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()
	CopyUpstreamResponse(w, resp)
}

func (h *CartHandler) GetCartMe(w http.ResponseWriter, r *http.Request) {
	userId := middleware.GetUserID(r.Context())
	resp, err := h.c.GetCart(r.Context(), userId, r.URL.RawQuery, r.Header)
	if err != nil {
		WriteUpstreamError(w, r, http.StatusBadGateway, "cart-service request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()
	CopyUpstreamResponse(w, resp)
}

func (h *CartHandler) CheckoutMe(w http.ResponseWriter, r *http.Request) {
	userId := middleware.GetUserID(r.Context())
	resp, err := h.c.Checkout(r.Context(), userId, r.URL.RawQuery, r.Body, r.Header)
	if err != nil {
		WriteUpstreamError(w, r, http.StatusBadGateway, "cart-service request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()
	CopyUpstreamResponse(w, resp)
}
