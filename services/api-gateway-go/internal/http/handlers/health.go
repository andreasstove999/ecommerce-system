package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/clients"
)

type HealthHandler struct {
	Probes []clients.HealthProbe
}

func (h *HealthHandler) Gateway(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "api-gateway",
	})
}

func (h *HealthHandler) Upstreams(w http.ResponseWriter, r *http.Request) {
	results := make([]clients.HealthResult, len(h.Probes))

	var wg sync.WaitGroup
	wg.Add(len(h.Probes))
	for i := range h.Probes {
		i := i
		go func() {
			defer wg.Done()
			results[i] = clients.CheckHealth(r.Context(), h.Probes[i])
		}()
	}
	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "ok",
		"service":  "api-gateway",
		"upstream": results,
	})
}
