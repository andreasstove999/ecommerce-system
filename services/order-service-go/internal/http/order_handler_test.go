package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	createFunc            func(ctx context.Context, o *order.Order) error
	getByIDFunc           func(ctx context.Context, orderID string) (*order.Order, error)
	listByUserFunc        func(ctx context.Context, userID string) ([]order.Order, error)
	markPaymentSucceeded  func(ctx context.Context, orderID string) (*order.CompletionState, error)
	markPaymentFailed     func(ctx context.Context, orderID string, reason string) error
	markStockReservedFunc func(ctx context.Context, orderID string) (*order.CompletionState, error)
	markCompletedFunc     func(ctx context.Context, orderID string) error
}

func (f *fakeRepo) Create(ctx context.Context, o *order.Order) error {
	if f.createFunc != nil {
		return f.createFunc(ctx, o)
	}
	return nil
}

func (f *fakeRepo) CreateWithTx(ctx context.Context, tx *sql.Tx, o *order.Order) error {
	return f.Create(ctx, o)
}

func (f *fakeRepo) GetByID(ctx context.Context, orderID string) (*order.Order, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(ctx, orderID)
	}
	return nil, nil
}

func (f *fakeRepo) ListByUser(ctx context.Context, userID string) ([]order.Order, error) {
	if f.listByUserFunc != nil {
		return f.listByUserFunc(ctx, userID)
	}
	return nil, nil
}

func (f *fakeRepo) MarkPaymentSucceeded(ctx context.Context, orderID string) (*order.CompletionState, error) {
	if f.markPaymentSucceeded != nil {
		return f.markPaymentSucceeded(ctx, orderID)
	}
	return nil, nil
}

func (f *fakeRepo) MarkPaymentFailed(ctx context.Context, orderID string, reason string) error {
	if f.markPaymentFailed != nil {
		return f.markPaymentFailed(ctx, orderID, reason)
	}
	return nil
}

func (f *fakeRepo) MarkStockReserved(ctx context.Context, orderID string) (*order.CompletionState, error) {
	if f.markStockReservedFunc != nil {
		return f.markStockReservedFunc(ctx, orderID)
	}
	return nil, nil
}

func (f *fakeRepo) MarkCompleted(ctx context.Context, orderID string) error {
	if f.markCompletedFunc != nil {
		return f.markCompletedFunc(ctx, orderID)
	}
	return nil
}

func TestGetOrder_Success(t *testing.T) {
	repo := &fakeRepo{
		getByIDFunc: func(ctx context.Context, orderID string) (*order.Order, error) {
			return &order.Order{
				ID:          orderID,
				CartID:      "cart-1",
				UserID:      "user-1",
				TotalAmount: 50,
				CreatedAt:   time.Unix(0, 0),
			}, nil
		},
	}
	handler := NewOrderHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/orders/abc", nil)
	req.SetPathValue("orderId", "abc")
	rr := httptest.NewRecorder()

	handler.GetOrder(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp order.Order
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "abc", resp.ID)
	assert.Equal(t, "user-1", resp.UserID)
}

func TestGetOrder_MissingPathParam(t *testing.T) {
	handler := NewOrderHandler(&fakeRepo{})

	req := httptest.NewRequest(http.MethodGet, "/api/orders/", nil)
	rr := httptest.NewRecorder()

	handler.GetOrder(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "missing orderId", resp["error"])
}

func TestGetOrder_RepositoryError(t *testing.T) {
	repo := &fakeRepo{
		getByIDFunc: func(ctx context.Context, orderID string) (*order.Order, error) {
			return nil, errors.New("db down")
		},
	}
	handler := NewOrderHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/orders/abc", nil)
	req.SetPathValue("orderId", "abc")
	rr := httptest.NewRecorder()

	handler.GetOrder(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGetOrder_NotFound(t *testing.T) {
	repo := &fakeRepo{
		getByIDFunc: func(ctx context.Context, orderID string) (*order.Order, error) {
			return nil, nil
		},
	}
	handler := NewOrderHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/orders/abc", nil)
	req.SetPathValue("orderId", "abc")
	rr := httptest.NewRecorder()

	handler.GetOrder(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "order not found", resp["error"])
}

func TestListOrdersByUser_Success(t *testing.T) {
	repo := &fakeRepo{
		listByUserFunc: func(ctx context.Context, userID string) ([]order.Order, error) {
			return []order.Order{
				{ID: "o1", UserID: userID},
				{ID: "o2", UserID: userID},
			}, nil
		},
	}
	handler := NewOrderHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/users/user-123/orders", nil)
	req.SetPathValue("userId", "user-123")
	rr := httptest.NewRecorder()

	handler.ListOrdersByUser(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp []order.Order
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Len(t, resp, 2)
	assert.Equal(t, "o1", resp[0].ID)
}

func TestListOrdersByUser_MissingUser(t *testing.T) {
	handler := NewOrderHandler(&fakeRepo{})

	req := httptest.NewRequest(http.MethodGet, "/api/users//orders", nil)
	rr := httptest.NewRecorder()

	handler.ListOrdersByUser(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "missing userId", resp["error"])
}

func TestListOrdersByUser_RepositoryError(t *testing.T) {
	repo := &fakeRepo{
		listByUserFunc: func(ctx context.Context, userID string) ([]order.Order, error) {
			return nil, errors.New("oops")
		},
	}
	handler := NewOrderHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/users/user-err/orders", nil)
	req.SetPathValue("userId", "user-err")
	rr := httptest.NewRecorder()

	handler.ListOrdersByUser(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestListOrdersByUser_EmptyList(t *testing.T) {
	repo := &fakeRepo{
		listByUserFunc: func(ctx context.Context, userID string) ([]order.Order, error) {
			return []order.Order{}, nil
		},
	}
	handler := NewOrderHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/users/user-empty/orders", nil)
	req.SetPathValue("userId", "user-empty")
	rr := httptest.NewRecorder()

	handler.ListOrdersByUser(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp []order.Order
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Empty(t, resp)
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	healthHandler(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "ok", resp["status"])
	assert.Equal(t, "order-service", resp["service"])
}
