package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	httpserver "github.com/andreasstove999/ecommerce-system/order-service-go/internal/http"
	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/testutil"
)

func TestGET_OrderByID_Returns200(t *testing.T) {
	db, cleanup := testutil.StartPostgres(t)
	t.Cleanup(cleanup)

	repo := order.NewRepository(db)

	seedCtx, seedCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer seedCancel()

	createdAt := time.Now().UTC().Truncate(time.Millisecond)
	seededOrder := order.Order{
		CartID:      "cart-http-123",
		UserID:      "user-http",
		TotalAmount: 64.75,
		CreatedAt:   createdAt,
		Items: []order.Item{
			{ProductID: "prod-1", Quantity: 1, Price: 14.75},
			{ProductID: "prod-2", Quantity: 2, Price: 25.00},
		},
	}

	require.NoError(t, repo.Create(seedCtx, &seededOrder))

	router := httpserver.NewRouter(repo)

	reqCtx, reqCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer reqCancel()

	req := httptest.NewRequest(http.MethodGet, "/api/orders/"+seededOrder.ID, nil).WithContext(reqCtx)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp order.Order
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))

	require.Equal(t, seededOrder.ID, resp.ID)
	require.Equal(t, seededOrder.UserID, resp.UserID)
	require.Equal(t, seededOrder.CartID, resp.CartID)
	require.Equal(t, seededOrder.TotalAmount, resp.TotalAmount)
	require.WithinDuration(t, seededOrder.CreatedAt, resp.CreatedAt, time.Millisecond)
	require.Len(t, resp.Items, len(seededOrder.Items))
}

func TestGET_OrderByID_NotFound_Returns404(t *testing.T) {
	db, cleanup := testutil.StartPostgres(t)
	t.Cleanup(cleanup)

	repo := order.NewRepository(db)
	router := httpserver.NewRouter(repo)

	reqCtx, reqCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer reqCancel()

	missingID := uuid.NewString()

	req := httptest.NewRequest(http.MethodGet, "/api/orders/"+missingID, nil).WithContext(reqCtx)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, "order not found", resp["error"])
}

func TestGET_ListOrdersByUser_Returns200(t *testing.T) {
	db, cleanup := testutil.StartPostgres(t)
	t.Cleanup(cleanup)

	repo := order.NewRepository(db)

	seedCtx, seedCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer seedCancel()

	userID := "user-list-http"
	now := time.Now().UTC().Truncate(time.Millisecond)

	firstOrder := order.Order{
		CartID:      "cart-1",
		UserID:      userID,
		TotalAmount: 30.00,
		CreatedAt:   now.Add(-1 * time.Minute),
		Items: []order.Item{
			{ProductID: "prod-a", Quantity: 1, Price: 10.00},
		},
	}

	secondOrder := order.Order{
		CartID:      "cart-2",
		UserID:      userID,
		TotalAmount: 50.00,
		CreatedAt:   now,
		Items: []order.Item{
			{ProductID: "prod-b", Quantity: 2, Price: 25.00},
		},
	}

	require.NoError(t, repo.Create(seedCtx, &firstOrder))
	require.NoError(t, repo.Create(seedCtx, &secondOrder))

	router := httpserver.NewRouter(repo)

	reqCtx, reqCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer reqCancel()

	req := httptest.NewRequest(http.MethodGet, "/api/users/"+userID+"/orders", nil).WithContext(reqCtx)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var orders []order.Order
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&orders))

	require.Len(t, orders, 2)
	require.Equal(t, userID, orders[0].UserID)
	require.Equal(t, userID, orders[1].UserID)
}
