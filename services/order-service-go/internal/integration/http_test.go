package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	httpserver "github.com/andreasstove999/ecommerce-system/order-service-go/internal/http"
	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/testutil"
)

func TestGET_OrderByID_Returns200(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	db, cleanup := testutil.StartPostgres(ctx, t)
	t.Cleanup(cleanup)

	repo := order.NewRepository(db)

	orderToCreate := &order.Order{
		ID:     uuid.NewString(),
		CartID: "cart-123",
		UserID: "user-abc",
		Items: []order.Item{
			{ProductID: "p1", Quantity: 2, Price: 12.50},
		},
		TotalAmount: 25.00,
		CreatedAt:   time.Now().UTC().Truncate(time.Millisecond),
	}

	createCtx, createCancel := context.WithTimeout(ctx, 5*time.Second)
	defer createCancel()

	if err := repo.Create(createCtx, orderToCreate); err != nil {
		t.Fatalf("seed order: %v", err)
	}

	router := httpserver.NewRouter(repo)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/orders/%s", orderToCreate.ID), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var got order.Order
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("parse body: %v", err)
	}

	if got.ID != orderToCreate.ID {
		t.Fatalf("expected orderId %s, got %s", orderToCreate.ID, got.ID)
	}
	if got.UserID == "" || got.CartID == "" || got.CreatedAt.IsZero() {
		t.Fatalf("missing required fields in response: %+v", got)
	}
	if len(got.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got.Items))
	}
	if got.TotalAmount != orderToCreate.TotalAmount {
		t.Fatalf("expected totalAmount %.2f, got %.2f", orderToCreate.TotalAmount, got.TotalAmount)
	}
}

func TestGET_OrderByID_NotFound_Returns404(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	db, cleanup := testutil.StartPostgres(ctx, t)
	t.Cleanup(cleanup)

	repo := order.NewRepository(db)
	router := httpserver.NewRouter(repo)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/orders/%s", uuid.NewString()), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("parse body: %v", err)
	}

	if body["error"] == "" {
		t.Fatalf("expected error message in body, got %+v", body)
	}
}

func TestGET_ListOrdersByUser_Returns200(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	db, cleanup := testutil.StartPostgres(ctx, t)
	t.Cleanup(cleanup)

	repo := order.NewRepository(db)

	userID := "user-list"

	for i := 0; i < 2; i++ {
		createCtx, createCancel := context.WithTimeout(ctx, 5*time.Second)
		if err := repo.Create(createCtx, &order.Order{
			ID:     uuid.NewString(),
			CartID: fmt.Sprintf("cart-%d", i),
			UserID: userID,
			Items: []order.Item{
				{ProductID: fmt.Sprintf("p-%d", i), Quantity: 1, Price: 10.00},
			},
			TotalAmount: 10.00,
			CreatedAt:   time.Now().UTC().Add(time.Duration(i) * time.Minute).Truncate(time.Millisecond),
		}); err != nil {
			createCancel()
			t.Fatalf("seed order %d: %v", i, err)
		}
		createCancel()
	}

	router := httpserver.NewRouter(repo)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%s/orders", userID), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var orders []order.Order
	if err := json.Unmarshal(rec.Body.Bytes(), &orders); err != nil {
		t.Fatalf("parse body: %v", err)
	}

	if len(orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(orders))
	}
}
