package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeEventRepo struct {
	createFunc              func(ctx context.Context, o *order.Order) error
	createWithTxFunc        func(ctx context.Context, tx *sql.Tx, o *order.Order) error
	markPaymentSucceeded    func(ctx context.Context, orderID string) (*order.CompletionState, error)
	markPaymentFailed       func(ctx context.Context, orderID string, reason string) error
	markStockReserved       func(ctx context.Context, orderID string) (*order.CompletionState, error)
	markCompleted           func(ctx context.Context, orderID string) error
	createdOrder            *order.Order
	markCompletedInvoked    bool
	markCompletedInvokedID  string
	markPaymentFailedCalled bool
	markPaymentFailedReason string
}

type fakeDedupRepo struct {
	last      int64
	found     bool
	getErr    error
	upsertErr error
	upserted  int64
	getCalls  int
	upsertTx  *sql.Tx
}

func (f *fakeDedupRepo) GetLastSequence(ctx context.Context, consumerName, partitionKey string) (int64, bool, error) {
	f.getCalls++
	return f.last, f.found, f.getErr
}

func (f *fakeDedupRepo) UpsertLastSequence(ctx context.Context, tx *sql.Tx, consumerName, partitionKey string, newSeq int64) error {
	f.upsertTx = tx
	f.upserted = newSeq
	return f.upsertErr
}

type fakePublisher struct {
	orderCreatedCalls   int
	orderCompletedCalls int
	lastMeta            EnvelopeMetadata
}

func (f *fakePublisher) PublishOrderCreated(ctx context.Context, o *order.Order, meta EnvelopeMetadata) error {
	f.orderCreatedCalls++
	f.lastMeta = meta
	return nil
}

func (f *fakePublisher) PublishOrderCompleted(ctx context.Context, orderID, userID string, meta EnvelopeMetadata) error {
	f.orderCompletedCalls++
	f.lastMeta = meta
	return nil
}

func (f *fakeEventRepo) Create(ctx context.Context, o *order.Order) error {
	f.createdOrder = o
	if f.createFunc != nil {
		return f.createFunc(ctx, o)
	}
	return nil
}

func (f *fakeEventRepo) CreateWithTx(ctx context.Context, tx *sql.Tx, o *order.Order) error {
	f.createdOrder = o
	if f.createWithTxFunc != nil {
		return f.createWithTxFunc(ctx, tx, o)
	}
	return nil
}

func (f *fakeEventRepo) GetByID(ctx context.Context, orderID string) (*order.Order, error) {
	return nil, nil
}

func (f *fakeEventRepo) ListByUser(ctx context.Context, userID string) ([]order.Order, error) {
	return nil, nil
}

func (f *fakeEventRepo) MarkPaymentSucceeded(ctx context.Context, orderID string) (*order.CompletionState, error) {
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

func (f *fakeEventRepo) MarkStockReserved(ctx context.Context, orderID string) (*order.CompletionState, error) {
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

	handler := CartCheckedOutHandler(nil, repo, &fakeDedupRepo{}, &fakePublisher{}, log.New(io.Discard, "", 0), false)

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
	handler := CartCheckedOutHandler(nil, repo, &fakeDedupRepo{}, &fakePublisher{}, log.New(io.Discard, "", 0), false)

	body := []byte(`{"cartId":"cart-1","userId":"user-1","items":[],"totalAmount":0,"timestamp":"2024-01-01T00:00:00Z"}`)

	err := handler(context.Background(), body)
	require.Error(t, err)
}

func TestHandlePaymentSucceeded_NotReady(t *testing.T) {
	repo := &fakeEventRepo{
		markPaymentSucceeded: func(ctx context.Context, orderID string) (*order.CompletionState, error) {
			return &order.CompletionState{
				UserID:          "user-1",
				ReadyToComplete: false,
			}, nil
		},
	}

	handler := PaymentSucceededHandler(repo, &fakePublisher{}, log.New(io.Discard, "", 0))

	body := []byte(`{"orderId":"order-1","userId":"user-1","timestamp":"2024-01-01T00:00:00Z"}`)
	require.NoError(t, handler(context.Background(), body))
	assert.False(t, repo.markCompletedInvoked)
}

func TestHandlePaymentSucceeded_ReadyButCompleteFails(t *testing.T) {
	repo := &fakeEventRepo{
		markPaymentSucceeded: func(ctx context.Context, orderID string) (*order.CompletionState, error) {
			return &order.CompletionState{
				UserID:          "user-1",
				ReadyToComplete: true,
			}, nil
		},
		markCompleted: func(ctx context.Context, orderID string) error {
			return errors.New("complete failed")
		},
	}

	handler := PaymentSucceededHandler(repo, &fakePublisher{}, log.New(io.Discard, "", 0))

	body := []byte(`{"orderId":"order-1","userId":"user-1","timestamp":"2024-01-01T00:00:00Z"}`)
	err := handler(context.Background(), body)
	require.Error(t, err)
	assert.True(t, repo.markCompletedInvoked)
	assert.Equal(t, "order-1", repo.markCompletedInvokedID)
}

func TestHandlePaymentSucceeded_Error(t *testing.T) {
	repo := &fakeEventRepo{
		markPaymentSucceeded: func(ctx context.Context, orderID string) (*order.CompletionState, error) {
			return nil, errors.New("update failed")
		},
	}
	handler := PaymentSucceededHandler(repo, &fakePublisher{}, log.New(io.Discard, "", 0))

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
		markStockReserved: func(ctx context.Context, orderID string) (*order.CompletionState, error) {
			return &order.CompletionState{
				UserID:          "user-1",
				ReadyToComplete: false,
			}, nil
		},
	}

	handler := StockReservedHandler(repo, &fakePublisher{}, log.New(io.Discard, "", 0))

	body := []byte(`{"orderId":"order-1","userId":"user-1","timestamp":"2024-01-01T00:00:00Z"}`)
	require.NoError(t, handler(context.Background(), body))
	assert.False(t, repo.markCompletedInvoked)
}

func TestHandleStockReserved_ReadyButCompleteFails(t *testing.T) {
	repo := &fakeEventRepo{
		markStockReserved: func(ctx context.Context, orderID string) (*order.CompletionState, error) {
			return &order.CompletionState{
				UserID:          "user-1",
				ReadyToComplete: true,
			}, nil
		},
		markCompleted: func(ctx context.Context, orderID string) error {
			return errors.New("complete failed")
		},
	}

	handler := StockReservedHandler(repo, &fakePublisher{}, log.New(io.Discard, "", 0))

	body := []byte(`{"orderId":"order-1","userId":"user-1","timestamp":"2024-01-01T00:00:00Z"}`)
	err := handler(context.Background(), body)
	require.Error(t, err)
	assert.True(t, repo.markCompletedInvoked)
	assert.Equal(t, "order-1", repo.markCompletedInvokedID)
}

func TestHandleStockReserved_Error(t *testing.T) {
	repo := &fakeEventRepo{
		markStockReserved: func(ctx context.Context, orderID string) (*order.CompletionState, error) {
			return nil, errors.New("update failed")
		},
	}
	handler := StockReservedHandler(repo, &fakePublisher{}, log.New(io.Discard, "", 0))

	body := []byte(`{"orderId":"order-1","userId":"user-1","timestamp":"2024-01-01T00:00:00Z"}`)
	err := handler(context.Background(), body)
	require.Error(t, err)
}

func TestCartCheckedOutHandler_DedupIgnoresDuplicateSequence(t *testing.T) {
	repo := &fakeEventRepo{}
	pub := &fakePublisher{}
	dedupRepo := &fakeDedupRepo{last: 2, found: true}

	handler := CartCheckedOutHandler(nil, repo, dedupRepo, pub, log.New(io.Discard, "", 0), true)

	seq := int64(2)
	env := CartCheckedOutEnvelope{
		EventName:    cartCheckedOutEventName,
		EventVersion: cartCheckedOutEventVersion,
		EventID:      "e1",
		PartitionKey: "cart-dup",
		Sequence:     &seq,
		Schema:       "contracts/events/cart/CartCheckedOut.v1.payload.schema.json",
		Payload: CartCheckedOutPayload{
			CartID:      "cart-dup",
			UserID:      "user-1",
			Items:       []CartItem{{ProductID: "p1", Quantity: 1, Price: 5}},
			TotalAmount: 5,
			Timestamp:   time.Now(),
		},
	}

	body, err := json.Marshal(env)
	require.NoError(t, err)

	require.NoError(t, handler(context.Background(), body))
	assert.Nil(t, repo.createdOrder)
	assert.Equal(t, 0, pub.orderCreatedCalls)
}

func TestCartCheckedOutHandler_DedupIgnoresLowerSequence(t *testing.T) {
	repo := &fakeEventRepo{}
	pub := &fakePublisher{}
	dedupRepo := &fakeDedupRepo{last: 5, found: true}

	handler := CartCheckedOutHandler(nil, repo, dedupRepo, pub, log.New(io.Discard, "", 0), true)

	seq := int64(3)
	env := CartCheckedOutEnvelope{
		EventName:    cartCheckedOutEventName,
		EventVersion: cartCheckedOutEventVersion,
		EventID:      "e1",
		PartitionKey: "cart-low",
		Sequence:     &seq,
		Schema:       "contracts/events/cart/CartCheckedOut.v1.payload.schema.json",
		Payload: CartCheckedOutPayload{
			CartID:      "cart-low",
			UserID:      "user-2",
			Items:       []CartItem{{ProductID: "p1", Quantity: 1, Price: 5}},
			TotalAmount: 5,
			Timestamp:   time.Now(),
		},
	}

	body, err := json.Marshal(env)
	require.NoError(t, err)

	require.NoError(t, handler(context.Background(), body))
	assert.Nil(t, repo.createdOrder)
	assert.Equal(t, 0, pub.orderCreatedCalls)
}

func TestCartCheckedOutHandler_DedupProcessesHigherSequence(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	repo := &fakeEventRepo{
		createWithTxFunc: func(ctx context.Context, tx *sql.Tx, o *order.Order) error {
			return nil
		},
	}
	pub := &fakePublisher{}
	dedupRepo := &fakeDedupRepo{last: 1, found: true}

	handler := CartCheckedOutHandler(db, repo, dedupRepo, pub, log.New(io.Discard, "", 0), true)

	seq := int64(3)
	env := CartCheckedOutEnvelope{
		EventName:     cartCheckedOutEventName,
		EventVersion:  cartCheckedOutEventVersion,
		EventID:       "e2",
		CorrelationID: "corr-1",
		PartitionKey:  "cart-gap",
		Sequence:      &seq,
		Schema:        "contracts/events/cart/CartCheckedOut.v1.payload.schema.json",
		Payload: CartCheckedOutPayload{
			CartID:      "cart-gap",
			UserID:      "user-3",
			Items:       []CartItem{{ProductID: "p1", Quantity: 1, Price: 5}},
			TotalAmount: 5,
			Timestamp:   time.Now(),
		},
	}

	body, err := json.Marshal(env)
	require.NoError(t, err)

	require.NoError(t, handler(context.Background(), body))
	assert.Equal(t, int64(3), dedupRepo.upserted)
	assert.Equal(t, 1, pub.orderCreatedCalls)
	assert.Equal(t, "corr-1", pub.lastMeta.CorrelationID)
	require.NoError(t, mock.ExpectationsWereMet())
}
