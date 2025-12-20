package events

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"

	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeEventRepo struct {
	createFunc              func(ctx context.Context, o *order.Order) error
	markPaymentSucceeded    func(ctx context.Context, orderID string) (*CompletionState, error)
	markPaymentFailed       func(ctx context.Context, orderID string, reason string) error
	markStockReserved       func(ctx context.Context, orderID string) (*CompletionState, error)
	markCompleted           func(ctx context.Context, orderID string) error
	createdOrder            *order.Order
	markCompletedInvoked    bool
	markCompletedInvokedID  string
	markPaymentFailedCalled bool
	markPaymentFailedReason string
}

func (f *fakeEventRepo) Create(ctx context.Context, o *order.Order) error {
	f.createdOrder = o
	if f.createFunc != nil {
		return f.createFunc(ctx, o)
	}
	return nil
}

func (f *fakeEventRepo) GetByID(ctx context.Context, orderID string) (*order.Order, error) {
	return nil, nil
}

func (f *fakeEventRepo) ListByUser(ctx context.Context, userID string) ([]order.Order, error) {
	return nil, nil
}

func (f *fakeEventRepo) MarkPaymentSucceeded(ctx context.Context, orderID string) (*CompletionState, error) {
	if f.markPaymentSucceeded != nil {
		return f.markPaymentSucceeded(ctx, orderID)
	}
	return nil, nil
}

func (f *fakeEventRepo) MarkPaymentFailed(ctx context.Context, orderID string, reason string) error {
	f.markPaymentFailedCalled = true
	f.markPaymentFailedReason = reason
	if f.markPaymentFailed != nil {
		return f.markPaymentFailed(ctx, orderID, reason)
	}
	return nil
}

func (f *fakeEventRepo) MarkStockReserved(ctx context.Context, orderID string) (*CompletionState, error) {
	if f.markStockReserved != nil {
		return f.markStockReserved(ctx, orderID)
	}
	return nil, nil
}

func (f *fakeEventRepo) MarkCompleted(ctx context.Context, orderID string) error {
	f.markCompletedInvoked = true
	f.markCompletedInvokedID = orderID
	if f.markCompleted != nil {
		return f.markCompleted(ctx, orderID)
	}
	return nil
}

func TestHandleCartCheckedOut_CreatesOrder(t *testing.T) {
	repo := &fakeEventRepo{
		createFunc: func(ctx context.Context, o *order.Order) error {
			// return an error to stop before publisher is used
			return errors.New("stop before publish")
		},
	}

	handler := CartCheckedOutHandler(repo, nil, log.New(io.Discard, "", 0))

	body := []byte(`{
		"eventType": "CartCheckedOut",
		"cartId": "cart-1",
		"userId": "user-1",
		"items": [{"productId": "p1", "quantity": 2, "price": 3.5}],
		"totalAmount": 7.0,
		"timestamp": "2024-01-01T00:00:00Z"
	}`)

	err := handler(context.Background(), body)
	require.Error(t, err)

	require.NotNil(t, repo.createdOrder)
	assert.Equal(t, "cart-1", repo.createdOrder.CartID)
	assert.Equal(t, "user-1", repo.createdOrder.UserID)
	require.Len(t, repo.createdOrder.Items, 1)
	assert.Equal(t, "p1", repo.createdOrder.Items[0].ProductID)
	assert.Equal(t, 2, repo.createdOrder.Items[0].Quantity)
}

func TestHandleCartCheckedOut_CreateError(t *testing.T) {
	repo := &fakeEventRepo{
		createFunc: func(ctx context.Context, o *order.Order) error {
			return errors.New("insert failed")
		},
	}
	handler := CartCheckedOutHandler(repo, nil, log.New(io.Discard, "", 0))

	body := []byte(`{"cartId":"cart-1","userId":"user-1","items":[],"totalAmount":0,"timestamp":"2024-01-01T00:00:00Z"}`)

	err := handler(context.Background(), body)
	require.Error(t, err)
}

func TestHandlePaymentSucceeded_NotReady(t *testing.T) {
	repo := &fakeEventRepo{
		markPaymentSucceeded: func(ctx context.Context, orderID string) (*CompletionState, error) {
			return &CompletionState{
				UserID:          "user-1",
				ReadyToComplete: false,
			}, nil
		},
	}

	handler := PaymentSucceededHandler(repo, nil, log.New(io.Discard, "", 0))

	body := []byte(`{"orderId":"order-1","userId":"user-1","timestamp":"2024-01-01T00:00:00Z"}`)
	require.NoError(t, handler(context.Background(), body))
	assert.False(t, repo.markCompletedInvoked)
}

func TestHandlePaymentSucceeded_ReadyButCompleteFails(t *testing.T) {
	repo := &fakeEventRepo{
		markPaymentSucceeded: func(ctx context.Context, orderID string) (*CompletionState, error) {
			return &CompletionState{
				UserID:          "user-1",
				ReadyToComplete: true,
			}, nil
		},
		markCompleted: func(ctx context.Context, orderID string) error {
			return errors.New("complete failed")
		},
	}

	handler := PaymentSucceededHandler(repo, nil, log.New(io.Discard, "", 0))

	body := []byte(`{"orderId":"order-1","userId":"user-1","timestamp":"2024-01-01T00:00:00Z"}`)
	err := handler(context.Background(), body)
	require.Error(t, err)
	assert.True(t, repo.markCompletedInvoked)
	assert.Equal(t, "order-1", repo.markCompletedInvokedID)
}

func TestHandlePaymentSucceeded_Error(t *testing.T) {
	repo := &fakeEventRepo{
		markPaymentSucceeded: func(ctx context.Context, orderID string) (*CompletionState, error) {
			return nil, errors.New("update failed")
		},
	}
	handler := PaymentSucceededHandler(repo, nil, log.New(io.Discard, "", 0))

	body := []byte(`{"orderId":"order-1","userId":"user-1","timestamp":"2024-01-01T00:00:00Z"}`)
	err := handler(context.Background(), body)
	require.Error(t, err)
}

func TestHandlePaymentFailed(t *testing.T) {
	repo := &fakeEventRepo{}
	handler := PaymentFailedHandler(repo, log.New(io.Discard, "", 0))

	body := []byte(`{"orderId":"order-1","reason":"declined","userId":"user-1","timestamp":"2024-01-01T00:00:00Z"}`)
	require.NoError(t, handler(context.Background(), body))

	assert.True(t, repo.markPaymentFailedCalled)
	assert.Equal(t, "declined", repo.markPaymentFailedReason)
}

func TestHandlePaymentFailed_Error(t *testing.T) {
	repo := &fakeEventRepo{
		markPaymentFailed: func(ctx context.Context, orderID string, reason string) error {
			return errors.New("update failed")
		},
	}
	handler := PaymentFailedHandler(repo, log.New(io.Discard, "", 0))

	body := []byte(`{"orderId":"order-1","reason":"declined","userId":"user-1","timestamp":"2024-01-01T00:00:00Z"}`)
	err := handler(context.Background(), body)
	require.Error(t, err)
}

func TestHandleStockReserved_NotReady(t *testing.T) {
	repo := &fakeEventRepo{
		markStockReserved: func(ctx context.Context, orderID string) (*CompletionState, error) {
			return &CompletionState{
				UserID:          "user-1",
				ReadyToComplete: false,
			}, nil
		},
	}

	handler := StockReservedHandler(repo, nil, log.New(io.Discard, "", 0))

	body := []byte(`{"orderId":"order-1","userId":"user-1","timestamp":"2024-01-01T00:00:00Z"}`)
	require.NoError(t, handler(context.Background(), body))
	assert.False(t, repo.markCompletedInvoked)
}

func TestHandleStockReserved_ReadyButCompleteFails(t *testing.T) {
	repo := &fakeEventRepo{
		markStockReserved: func(ctx context.Context, orderID string) (*CompletionState, error) {
			return &CompletionState{
				UserID:          "user-1",
				ReadyToComplete: true,
			}, nil
		},
		markCompleted: func(ctx context.Context, orderID string) error {
			return errors.New("complete failed")
		},
	}

	handler := StockReservedHandler(repo, nil, log.New(io.Discard, "", 0))

	body := []byte(`{"orderId":"order-1","userId":"user-1","timestamp":"2024-01-01T00:00:00Z"}`)
	err := handler(context.Background(), body)
	require.Error(t, err)
	assert.True(t, repo.markCompletedInvoked)
	assert.Equal(t, "order-1", repo.markCompletedInvokedID)
}

func TestHandleStockReserved_Error(t *testing.T) {
	repo := &fakeEventRepo{
		markStockReserved: func(ctx context.Context, orderID string) (*CompletionState, error) {
			return nil, errors.New("update failed")
		},
	}
	handler := StockReservedHandler(repo, nil, log.New(io.Discard, "", 0))

	body := []byte(`{"orderId":"order-1","userId":"user-1","timestamp":"2024-01-01T00:00:00Z"}`)
	err := handler(context.Background(), body)
	require.Error(t, err)
}
