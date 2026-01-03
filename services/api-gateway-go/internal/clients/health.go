package clients

import (
	"context"
	"net/http"
	"time"
)

type HealthProbe struct {
	Name   string
	Client *Client
	Path   string
}

type HealthResult struct {
	Name       string `json:"name"`
	OK         bool   `json:"ok"`
	StatusCode int    `json:"statusCode,omitempty"`
	Error      string `json:"error,omitempty"`
}

func CheckHealth(ctx context.Context, probe HealthProbe) HealthResult {
	// Short probe timeout
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := probe.Client.Do(ctx, http.MethodGet, probe.Path, "", nil, http.Header{})
	if err != nil {
		return HealthResult{Name: probe.Name, OK: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	ok := resp.StatusCode >= 200 && resp.StatusCode < 300
	return HealthResult{Name: probe.Name, OK: ok, StatusCode: resp.StatusCode}
}
