//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/http/dto"
)

type httpResult struct {
	StatusCode int
	Body       []byte
	Header     http.Header
}

func TestGatewayHappyPath(t *testing.T) {
	baseURL := getenv("GATEWAY_URL", "http://localhost:8080")
	userID := "integration-user"
	correlationID := fmt.Sprintf("it-%d", time.Now().UnixNano())

	client := &http.Client{Timeout: 10 * time.Second}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	waitForHealth(ctx, t, client, baseURL)

	productID := createProduct(ctx, t, client, baseURL, correlationID)
	adjustInventory(ctx, t, client, baseURL, productID, correlationID)
	addCartItem(ctx, t, client, baseURL, userID, productID, correlationID)
	checkout(ctx, t, client, baseURL, userID, correlationID)

	orderID := pollOrders(ctx, t, client, baseURL, userID, correlationID)
	if orderID == "" {
		t.Fatalf("expected an order to be created")
	}

	pollOptional(ctx, t, client, fmt.Sprintf("%s/orders/%s/payment", baseURL, orderID), correlationID)
	pollOptional(ctx, t, client, fmt.Sprintf("%s/orders/%s/shipping", baseURL, orderID), correlationID)
}

func createProduct(ctx context.Context, t *testing.T, client *http.Client, baseURL, cid string) string {
	payload := `{"sku":"sku-it","name":"integration product","description":"test","price":12.5,"currency":"USD"}`
	resp := doRequest(ctx, t, client, http.MethodPost, baseURL+"/products", payload, map[string]string{"X-Correlation-Id": cid})
	ensureNon5xx(t, resp)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status creating product: %d", resp.StatusCode)
	}

	var product dto.Product
	decodeJSON(t, resp.Body, &product)
	if product.ID == "" {
		t.Fatalf("missing product id in response: %+v", product)
	}
	return product.ID
}

func adjustInventory(ctx context.Context, t *testing.T, client *http.Client, baseURL, productID, cid string) {
	payload := fmt.Sprintf(`{"productId":"%s","available":5}`, productID)
	resp := doRequest(ctx, t, client, http.MethodPost, baseURL+"/inventory/adjust", payload, map[string]string{"X-Correlation-Id": cid})
	ensureNon5xx(t, resp)
}

func addCartItem(ctx context.Context, t *testing.T, client *http.Client, baseURL, userID, productID, cid string) {
	payload := fmt.Sprintf(`{"productId":"%s","quantity":1,"price":12.5}`, productID)
	headers := map[string]string{"X-User-Id": userID, "X-Correlation-Id": cid}
	resp := doRequest(ctx, t, client, http.MethodPost, baseURL+"/me/cart/items", payload, headers)
	ensureNon5xx(t, resp)
}

func checkout(ctx context.Context, t *testing.T, client *http.Client, baseURL, userID, cid string) {
	headers := map[string]string{"X-User-Id": userID, "X-Correlation-Id": cid}
	resp := doRequest(ctx, t, client, http.MethodPost, baseURL+"/me/cart/checkout", "{}", headers)
	ensureNon5xx(t, resp)
}

func pollOrders(ctx context.Context, t *testing.T, client *http.Client, baseURL, userID, cid string) string {
	headers := map[string]string{"X-User-Id": userID, "X-Correlation-Id": cid}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(20 * time.Second)

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("context done while polling orders: %v", ctx.Err())
		case <-timeout:
			t.Fatalf("timed out waiting for orders")
		case <-ticker.C:
			resp := doRequest(ctx, t, client, http.MethodGet, baseURL+"/me/orders", "", headers)
			ensureNon5xx(t, resp)

			var orders []dto.Order
			if err := json.Unmarshal(resp.Body, &orders); err != nil {
				t.Fatalf("failed to decode orders: %v", err)
			}
			if len(orders) == 0 {
				continue
			}
			if id := orders[0].OrderID; id != "" {
				return id
			}
		}
	}
}

func pollOptional(ctx context.Context, t *testing.T, client *http.Client, url, cid string) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for attempts := 0; attempts < 10; attempts++ {
		<-ticker.C
		resp := doRequest(ctx, t, client, http.MethodGet, url, "", map[string]string{"X-Correlation-Id": cid})
		if resp.StatusCode >= 500 {
			t.Fatalf("received 5xx from %s: %d", url, resp.StatusCode)
		}
		if resp.StatusCode == http.StatusOK {
			return
		}
	}
	t.Logf("resource %s not ready after polling", url)
}

func waitForHealth(ctx context.Context, t *testing.T, client *http.Client, baseURL string) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("context done while waiting for health: %v", ctx.Err())
		case <-ticker.C:
			// Check gateway health
			resp := doRequest(ctx, t, client, http.MethodGet, baseURL+"/health", "", nil)
			if resp.StatusCode != http.StatusOK {
				continue
			}
			// Check upstreams health
			resp = doRequest(ctx, t, client, http.MethodGet, baseURL+"/health/upstreams", "", nil)
			if resp.StatusCode != http.StatusOK {
				continue
			}

			// Optional: Parse body to ensure shipping-service is OK
			var payload struct {
				Upstream []struct {
					Name string `json:"name"`
					OK   bool   `json:"ok"`
				} `json:"upstream"`
			}
			if err := json.Unmarshal(resp.Body, &payload); err != nil {
				continue
			}

			// Verify shipping is present and OK
			shippingFound := false
			for _, u := range payload.Upstream {
				if u.Name == "shipping-service" && u.OK {
					shippingFound = true
					break
				}
			}
			if shippingFound {
				return
			}
		}
	}
}

func doRequest(ctx context.Context, t *testing.T, client *http.Client, method, url, body string, headers map[string]string) httpResult {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request %s %s: %v", method, url, err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	ensureCorrelation(t, resp.Header)

	return httpResult{StatusCode: resp.StatusCode, Body: data, Header: resp.Header.Clone()}
}

func ensureCorrelation(t *testing.T, h http.Header) {
	if h.Get("X-Correlation-Id") == "" {
		t.Fatalf("expected correlation id on response")
	}
}

func ensureNon5xx(t *testing.T, resp httpResult) {
	if resp.StatusCode >= 500 {
		t.Fatalf("received 5xx (%d): %s", resp.StatusCode, string(resp.Body))
	}
}

func decodeJSON(t *testing.T, b []byte, v any) {
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatalf("decode json: %v", err)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); strings.TrimSpace(v) != "" {
		return v
	}
	return def
}
