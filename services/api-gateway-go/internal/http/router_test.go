package http

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/clients"
	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/config"
	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/middleware"
)

type recordedRequest struct {
	Method   string
	Path     string
	RawQuery string
	Header   http.Header
	Body     string
}

func newStubServer(t *testing.T) (*httptest.Server, <-chan recordedRequest) {
	t.Helper()
	ch := make(chan recordedRequest, 10)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		ch <- recordedRequest{
			Method:   r.Method,
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
			Header:   r.Header.Clone(),
			Body:     string(body),
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	return srv, ch
}

func newRouterWithBaseURL(baseURL string) http.Handler {
	logger := log.New(io.Discard, "", 0)
	httpClient := &http.Client{Timeout: 5 * time.Second}
	base := clients.NewClient("test", baseURL, httpClient)

	return NewRouter(Deps{
		Logger:    logger,
		Cfg:       config.Config{CORSAllowOrigins: []string{"*"}},
		Cart:      clients.NewCartClient(base),
		Order:     clients.NewOrderClient(base),
		Inventory: clients.NewInventoryClient(base),
		Catalog:   clients.NewCatalogClient(base),
		Payment:   clients.NewPaymentClient(base),
		Shipping:  clients.NewShippingClient(base),
	})
}

func TestHealthRoute(t *testing.T) {
	router := newRouterWithBaseURL("http://example.com")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body["status"] != "ok" || body["service"] != "api-gateway" {
		t.Fatalf("unexpected body: %v", body)
	}
}

func TestRequireUserIDMiddleware(t *testing.T) {
	router := newRouterWithBaseURL("http://example.com")

	req := httptest.NewRequest(http.MethodGet, "/me/cart", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["error"] == nil {
		t.Fatalf("expected error message in response: %v", resp)
	}
}

func TestCorrelationIDEchoAndGeneration(t *testing.T) {
	router := newRouterWithBaseURL("http://example.com")

	reqWith := httptest.NewRequest(http.MethodGet, "/health", nil)
	reqWith.Header.Set("X-Correlation-Id", "abc")
	rrWith := httptest.NewRecorder()
	router.ServeHTTP(rrWith, reqWith)
	if got := rrWith.Header().Get("X-Correlation-Id"); got != "abc" {
		t.Fatalf("expected correlation id to be echoed, got %q", got)
	}

	reqGen := httptest.NewRequest(http.MethodGet, "/health", nil)
	rrGen := httptest.NewRecorder()
	router.ServeHTTP(rrGen, reqGen)
	if cid := rrGen.Header().Get("X-Correlation-Id"); cid == "" {
		t.Fatalf("expected generated correlation id to be present")
	}
}

func TestCORSPreflight(t *testing.T) {
	router := newRouterWithBaseURL("http://example.com")

	req := httptest.NewRequest(http.MethodOptions, "/products", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for preflight, got %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Fatalf("expected Access-Control-Allow-Origin header")
	}
	if rr.Header().Get("Access-Control-Allow-Methods") == "" || rr.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Fatalf("expected CORS allow headers to be set")
	}
}

func TestRecoverMiddleware(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	handler := middlewareRecover(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body["error"] != "internal server error" {
		t.Fatalf("unexpected error message: %v", body)
	}
}

func middlewareRecover(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return middleware.Recover(logger)(next)
	}
}

func TestForwardingForCatalogProducts(t *testing.T) {
	srv, ch := newStubServer(t)
	defer srv.Close()

	router := newRouterWithBaseURL(srv.URL)

	req := httptest.NewRequest(http.MethodGet, "/products?limit=10", nil)
	req.Header.Set("X-Correlation-Id", "cid-123")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	select {
	case rec := <-ch:
		if rec.Method != http.MethodGet || rec.Path != "/api/catalog/products" || rec.RawQuery != "limit=10" {
			t.Fatalf("unexpected upstream request: %+v", rec)
		}
		if rec.Header.Get("X-Correlation-Id") != "cid-123" {
			t.Fatalf("correlation id not forwarded: %v", rec.Header)
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive upstream request")
	}
}

func TestForwardingPathsAndMethods(t *testing.T) {
	srv, ch := newStubServer(t)
	defer srv.Close()

	router := newRouterWithBaseURL(srv.URL)

	cases := []struct {
		name     string
		method   string
		path     string
		headers  map[string]string
		wantPath string
	}{
		{name: "product by id", method: http.MethodGet, path: "/products/123", wantPath: "/api/catalog/products/123"},
		{name: "order by id", method: http.MethodGet, path: "/orders/ord-1", wantPath: "/api/orders/ord-1"},
		{name: "availability", method: http.MethodGet, path: "/products/sku-1/availability", wantPath: "/api/inventory/sku-1"},
		{name: "me orders", method: http.MethodGet, path: "/me/orders", wantPath: "/api/users/u-9/orders", headers: map[string]string{"X-User-Id": "u-9"}},
		{name: "add to cart", method: http.MethodPost, path: "/me/cart/items", wantPath: "/api/cart/u-9/items", headers: map[string]string{"X-User-Id": "u-9"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK && rr.Code != http.StatusNoContent {
				t.Fatalf("unexpected status: %d", rr.Code)
			}

			select {
			case rec := <-ch:
				if rec.Path != tc.wantPath {
					t.Fatalf("expected upstream path %s, got %s", tc.wantPath, rec.Path)
				}
				if rec.Method != tc.method {
					t.Fatalf("expected method %s, got %s", tc.method, rec.Method)
				}
				if cid := rec.Header.Get("X-Correlation-Id"); cid == "" {
					t.Fatalf("expected correlation id forwarded")
				}
			case <-time.After(time.Second):
				t.Fatal("did not receive upstream request")
			}
		})
	}
}

func TestForwardingBodyHeadersAndHopByHopStripping(t *testing.T) {
	srv, ch := newStubServer(t)
	defer srv.Close()

	router := newRouterWithBaseURL(srv.URL)

	body := strings.NewReader(`{"name":"item"}`)
	req := httptest.NewRequest(http.MethodPost, "/products", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Transfer-Encoding", "chunked")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	select {
	case rec := <-ch:
		if rec.Method != http.MethodPost || rec.Path != "/api/catalog/products" {
			t.Fatalf("unexpected upstream request: %+v", rec)
		}
		if got := rec.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("content type not forwarded, got %q", got)
		}
		if strings.TrimSpace(rec.Body) != `{"name":"item"}` {
			t.Fatalf("unexpected body: %q", rec.Body)
		}
		if rec.Header.Get("Connection") != "" || rec.Header.Get("Transfer-Encoding") != "" {
			t.Fatalf("hop-by-hop headers should not be forwarded: %v", rec.Header)
		}
		if rec.Header.Get("X-Correlation-Id") == "" {
			t.Fatalf("expected generated correlation id to be forwarded")
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive upstream request")
	}
}
