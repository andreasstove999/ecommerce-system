package handlers

import (
	"net/http"

	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/clients"
)

type CatalogHandler struct{ c *clients.CatalogClient }

func NewCatalogHandler(c *clients.CatalogClient) *CatalogHandler { return &CatalogHandler{c: c} }

func (h *CatalogHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	resp, err := h.c.ListProducts(r.Context(), r.URL.RawQuery, r.Header)
	if err != nil {
		WriteUpstreamError(w, r, http.StatusBadGateway, "catalog-service request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()
	CopyUpstreamResponse(w, resp)
}

func (h *CatalogHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	resp, err := h.c.GetProduct(r.Context(), id, r.URL.RawQuery, r.Header)
	if err != nil {
		WriteUpstreamError(w, r, http.StatusBadGateway, "catalog-service request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()
	CopyUpstreamResponse(w, resp)
}

func (h *CatalogHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	resp, err := h.c.CreateProduct(r.Context(), r.URL.RawQuery, r.Body, r.Header)
	if err != nil {
		WriteUpstreamError(w, r, http.StatusBadGateway, "catalog-service request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()
	CopyUpstreamResponse(w, resp)
}
