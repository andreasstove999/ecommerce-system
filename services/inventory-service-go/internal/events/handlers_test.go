package events

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/inventory"
)

type fakeInventory struct {
	orderID string
	lines   []inventory.Line
	result  inventory.ReserveResult
	err     error
	called  bool
}

func (f *fakeInventory) Get(ctx context.Context, productID string) (inventory.StockItem, error) {
	return inventory.StockItem{}, errors.New("not implemented")
}

func (f *fakeInventory) SetAvailable(ctx context.Context, productID string, available int) error {
	return errors.New("not implemented")
}

func (f *fakeInventory) Reserve(ctx context.Context, orderID string, lines []inventory.Line) (inventory.ReserveResult, error) {
	f.called = true
	f.orderID = orderID
	f.lines = append([]inventory.Line(nil), lines...)
	if f.err != nil {
		return inventory.ReserveResult{}, f.err
	}
	return f.result, nil
}

type fakePublisher struct {
	reservedQueue  string
	reservedBody   []byte
	reservedCalled bool
	reservedErr    error

	depletedQueue  string
	depletedBody   []byte
	depletedCalled bool
	depletedErr    error
}

func (f *fakePublisher) PublishStockReserved(ctx context.Context, orderID, userID string, reserved []inventory.Line) error {
	f.reservedCalled = true
	f.reservedQueue = StockReservedQueue
	if f.reservedErr != nil {
		return f.reservedErr
	}

	ev := StockReserved{
		EventType: EventTypeStockReserved,
		OrderID:   orderID,
		UserID:    userID,
		Timestamp: time.Unix(0, 0).UTC(),
	}
	for _, it := range reserved {
		ev.Items = append(ev.Items, StockLine{ProductID: it.ProductID, Quantity: it.Quantity})
	}

	body, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	f.reservedBody = body
	return nil
}

func (f *fakePublisher) PublishStockDepleted(ctx context.Context, orderID, userID string, depleted []inventory.DepletedLine, reserved []inventory.Line) error {
	f.depletedCalled = true
	f.depletedQueue = StockDepletedQueue
	if f.depletedErr != nil {
		return f.depletedErr
	}

	ev := StockDepleted{
		EventType: EventTypeStockDepleted,
		OrderID:   orderID,
		UserID:    userID,
		Timestamp: time.Unix(0, 0).UTC(),
	}
	for _, d := range depleted {
		ev.Depleted = append(ev.Depleted, DepletedLine{
			ProductID: d.ProductID,
			Requested: d.Requested,
			Available: d.Available,
		})
	}
	for _, r := range reserved {
		ev.Reserved = append(ev.Reserved, StockLine{
			ProductID: r.ProductID,
			Quantity:  r.Quantity,
		})
	}

	body, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	f.depletedBody = body
	return nil
}

func TestOrderCreatedHandler(t *testing.T) {
	t.Parallel()

	validEvent := OrderCreated{
		EventType: EventTypeOrderCreated,
		OrderID:   "order-123",
		UserID:    "user-42",
		Items: []CartItem{
			{ProductID: "p1", Quantity: 2},
			{ProductID: "p2", Quantity: 1},
			{ProductID: "", Quantity: 4},   // ignored
			{ProductID: "p3", Quantity: 0}, // ignored
		},
		Timestamp: time.Unix(0, 0).UTC(),
	}

	type tc struct {
		name       string
		body       []byte
		repo       *fakeInventory
		pub        *fakePublisher
		wantErr    bool
		wantRepo   bool
		assertFunc func(t *testing.T, repo *fakeInventory, pub *fakePublisher)
	}

	tests := []tc{
		{
			name: "reserves stock and publishes stock reserved",
			body: mustMarshal(validEvent, t),
			repo: &fakeInventory{
				result: inventory.ReserveResult{
					Reserved: []inventory.Line{{ProductID: "p1", Quantity: 2}, {ProductID: "p2", Quantity: 1}},
				},
			},
			pub:      &fakePublisher{},
			wantRepo: true,
			assertFunc: func(t *testing.T, repo *fakeInventory, pub *fakePublisher) {
				t.Helper()
				if !reflect.DeepEqual(repo.lines, []inventory.Line{{ProductID: "p1", Quantity: 2}, {ProductID: "p2", Quantity: 1}}) {
					t.Fatalf("Reserve called with %+v, want p1 and p2", repo.lines)
				}
				if !pub.reservedCalled {
					t.Fatalf("PublishStockReserved not called")
				}
				if pub.reservedQueue != StockReservedQueue {
					t.Fatalf("reserved queue=%s want=%s", pub.reservedQueue, StockReservedQueue)
				}

				var ev StockReserved
				if err := json.Unmarshal(pub.reservedBody, &ev); err != nil {
					t.Fatalf("reserved payload not JSON: %v", err)
				}
				if ev.OrderID != validEvent.OrderID || ev.UserID != validEvent.UserID || ev.EventType != EventTypeStockReserved {
					t.Fatalf("reserved payload mismatch: %+v", ev)
				}
				if !reflect.DeepEqual(ev.Items, []StockLine{{ProductID: "p1", Quantity: 2}, {ProductID: "p2", Quantity: 1}}) {
					t.Fatalf("reserved items=%+v", ev.Items)
				}
				if pub.depletedCalled {
					t.Fatalf("PublishStockDepleted should not be called")
				}
			},
		},
		{
			name: "publishes stock depleted on insufficient stock",
			body: mustMarshal(validEvent, t),
			repo: &fakeInventory{
				result: inventory.ReserveResult{
					Reserved: []inventory.Line{{ProductID: "p2", Quantity: 1}},
					Depleted: []inventory.DepletedLine{{ProductID: "p1", Requested: 2, Available: 1}},
				},
			},
			pub:      &fakePublisher{},
			wantRepo: true,
			assertFunc: func(t *testing.T, repo *fakeInventory, pub *fakePublisher) {
				t.Helper()
				if !pub.depletedCalled {
					t.Fatalf("PublishStockDepleted not called")
				}
				if pub.depletedQueue != StockDepletedQueue {
					t.Fatalf("depleted queue=%s want=%s", pub.depletedQueue, StockDepletedQueue)
				}

				var ev StockDepleted
				if err := json.Unmarshal(pub.depletedBody, &ev); err != nil {
					t.Fatalf("depleted payload not JSON: %v", err)
				}
				if ev.OrderID != validEvent.OrderID || ev.UserID != validEvent.UserID || ev.EventType != EventTypeStockDepleted {
					t.Fatalf("depleted payload mismatch: %+v", ev)
				}
				wantDepleted := []DepletedLine{{ProductID: "p1", Requested: 2, Available: 1}}
				if !reflect.DeepEqual(ev.Depleted, wantDepleted) {
					t.Fatalf("depleted lines=%+v want=%+v", ev.Depleted, wantDepleted)
				}
				wantReserved := []StockLine{{ProductID: "p2", Quantity: 1}}
				if !reflect.DeepEqual(ev.Reserved, wantReserved) {
					t.Fatalf("reserved lines=%+v want=%+v", ev.Reserved, wantReserved)
				}
				if pub.reservedCalled {
					t.Fatalf("PublishStockReserved should not be called")
				}
			},
		},
		{
			name:    "returns error on invalid JSON and does not publish",
			body:    []byte(`{"orderId":`),
			repo:    &fakeInventory{},
			pub:     &fakePublisher{},
			wantErr: true,
			assertFunc: func(t *testing.T, repo *fakeInventory, pub *fakePublisher) {
				t.Helper()
				if repo.called {
					t.Fatalf("Reserve should not be called on invalid JSON")
				}
				if pub.reservedCalled || pub.depletedCalled {
					t.Fatalf("publisher should not be called on invalid JSON")
				}
			},
		},
		{
			name: "returns error on publisher failure",
			body: mustMarshal(validEvent, t),
			repo: &fakeInventory{
				result: inventory.ReserveResult{
					Reserved: []inventory.Line{{ProductID: "p1", Quantity: 2}},
				},
			},
			pub:      &fakePublisher{reservedErr: errors.New("publish failed")},
			wantErr:  true,
			wantRepo: true,
			assertFunc: func(t *testing.T, repo *fakeInventory, pub *fakePublisher) {
				t.Helper()
				if !pub.reservedCalled {
					t.Fatalf("PublishStockReserved should be called even when returning error")
				}
			},
		},
		{
			name: "returns error on repository failure",
			body: mustMarshal(validEvent, t),
			repo: &fakeInventory{
				err: errors.New("db down"),
			},
			pub:     &fakePublisher{},
			wantErr: true,
			assertFunc: func(t *testing.T, repo *fakeInventory, pub *fakePublisher) {
				t.Helper()
				if pub.reservedCalled || pub.depletedCalled {
					t.Fatalf("publisher should not be called on repository error")
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := OrderCreatedHandler(tt.repo, tt.pub, log.New(io.Discard, "", 0))
			err := handler(context.Background(), tt.body)

			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantRepo && !tt.repo.called {
				t.Fatalf("Reserve was not called")
			}
			tt.assertFunc(t, tt.repo, tt.pub)
		})
	}
}

func mustMarshal(ev OrderCreated, t *testing.T) []byte {
	t.Helper()
	body, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	return body
}
