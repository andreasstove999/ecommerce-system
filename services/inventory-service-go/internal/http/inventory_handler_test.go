package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/inventory"
)

type fakeRepo struct{ items map[string]int }

func (r *fakeRepo) Get(ctx context.Context, productID string) (inventory.StockItem, error) {
	v, ok := r.items[productID]
	if !ok {
		return inventory.StockItem{}, inventory.ErrNotFound
	}
	return inventory.StockItem{ProductID: productID, Available: v}, nil
}
func (r *fakeRepo) SetAvailable(ctx context.Context, productID string, available int) error {
	if r.items == nil {
		r.items = map[string]int{}
	}
	r.items[productID] = available
	return nil
}
func (r *fakeRepo) Reserve(ctx context.Context, orderID string, lines []inventory.Line) (inventory.ReserveResult, error) {
	return inventory.ReserveResult{}, nil
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
}
