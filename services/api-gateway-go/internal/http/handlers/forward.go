package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/middleware"
	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/model"
)

func CopyUpstreamResponse(w http.ResponseWriter, resp *http.Response) {
	// Copy headers (avoid hop-by-hop)
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func WriteUpstreamError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(model.ErrorResponse{
		Error:         msg,
		CorrelationID: middleware.GetCorrelationID(r.Context()),
	})
}
