package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/inventory"
)

type fakeRepo struct {
	items  map[string]int
	getErr error
	setErr error
}

func (r *fakeRepo) Get(ctx context.Context, productID string) (inventory.StockItem, error) {
	if r.getErr != nil {
		return inventory.StockItem{}, r.getErr
	}
	v, ok := r.items[productID]
	if !ok {
		return inventory.StockItem{}, inventory.ErrNotFound
	}
	return inventory.StockItem{ProductID: productID, Available: v}, nil
}
func (r *fakeRepo) SetAvailable(ctx context.Context, productID string, available int) error {
	if r.setErr != nil {
		return r.setErr
	}
	if r.items == nil {
		r.items = map[string]int{}
	}
	r.items[productID] = available
	return nil
}
func (r *fakeRepo) Reserve(ctx context.Context, orderID string, lines []inventory.Line) (inventory.ReserveResult, error) {
	return inventory.ReserveResult{}, nil
}

func TestHealth(t *testing.T) {
	h := NewHandler(&fakeRepo{items: map[string]int{}})
	router := NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "ok" {
		t.Fatalf("expected body \"ok\", got %q", body)
	}
}

func TestGetAvailability_NotFound(t *testing.T) {
	repo := &fakeRepo{items: map[string]int{}}
	h := NewHandler(repo)
	r := NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/inventory/does-not-exist", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.Code)
	}
}

func TestGetAvailability_OK(t *testing.T) {
	repo := &fakeRepo{items: map[string]int{"p1": 3}}
	h := NewHandler(repo)
	r := NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/inventory/p1", nil)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	if ct := res.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("expected JSON content type, got %q", ct)
	}

	var item inventory.StockItem
	if err := json.NewDecoder(res.Body).Decode(&item); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if item.ProductID != "p1" || item.Available != 3 {
		t.Fatalf("unexpected body: %+v", item)
	}
}

func TestAdjustAvailability_OK(t *testing.T) {
	repo := &fakeRepo{items: map[string]int{}}
	h := NewHandler(repo)
	r := NewRouter(h)

	body := bytes.NewBufferString(`{"productId":"p1","available":7}`)
	req := httptest.NewRequest(http.MethodPost, "/api/inventory/adjust", body)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}

	if got := repo.items["p1"]; got != 7 {
		t.Fatalf("expected repo to store 7, got %d", got)
	}
}

func TestAdjustAvailability_StringValue(t *testing.T) {
	repo := &fakeRepo{items: map[string]int{}}
	h := NewHandler(repo)
	r := NewRouter(h)

	// Sending "7" as a string instead of a number
	body := bytes.NewBufferString(`{"productId":"p1","available":"7"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/inventory/adjust", body)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", res.Code, res.Body.String())
	}

	if got := repo.items["p1"]; got != 7 {
		t.Fatalf("expected repo to store 7, got %d", got)
	}
}

func TestAdjustAvailability_InvalidJSON(t *testing.T) {
	h := NewHandler(&fakeRepo{items: map[string]int{}})
	r := NewRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/api/inventory/adjust", strings.NewReader(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}

func TestAdjustAvailability_ServiceError(t *testing.T) {
	repo := &fakeRepo{items: map[string]int{}, setErr: errors.New("boom")}
	h := NewHandler(repo)
	r := NewRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/api/inventory/adjust", strings.NewReader(`{"productId":"p1","available":2}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", res.Code)
	}
}
