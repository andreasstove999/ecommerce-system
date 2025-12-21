package inventory

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fakeRepository struct {
	stock map[string]int

	reserveErr error
	reserveCnt int
}

func newFakeRepository(initial map[string]int) *fakeRepository {
	cp := make(map[string]int, len(initial))
	for k, v := range initial {
		cp[k] = v
	}
	return &fakeRepository{stock: cp}
}

func (f *fakeRepository) Get(ctx context.Context, productID string) (StockItem, error) {
	if v, ok := f.stock[productID]; ok {
		return StockItem{ProductID: productID, Available: v}, nil
	}
	return StockItem{}, ErrNotFound
}

func (f *fakeRepository) SetAvailable(ctx context.Context, productID string, available int) error {
	f.stock[productID] = available
	return nil
}

func (f *fakeRepository) Reserve(ctx context.Context, orderID string, lines []Line) (ReserveResult, error) {
	f.reserveCnt++
	if f.reserveErr != nil {
		return ReserveResult{}, f.reserveErr
	}

	var res ReserveResult
	for _, ln := range lines {
		available := f.stock[ln.ProductID]
		if available < ln.Quantity {
			res.Depleted = append(res.Depleted, DepletedLine{
				ProductID: ln.ProductID, Requested: ln.Quantity, Available: available,
			})
		}
	}
	if len(res.Depleted) > 0 {
		return res, nil
	}

	for _, ln := range lines {
		f.stock[ln.ProductID] -= ln.Quantity
		res.Reserved = append(res.Reserved, ln)
	}
	return res, nil
}

func TestServiceReserveForOrder(t *testing.T) {
	tests := map[string]struct {
		initialStock map[string]int
		lines        []Line
		repoErr      error
		orderID      string
		want         ReserveResult
		wantErr      bool
		wantReserve  int
	}{
		"all items available": {
			initialStock: map[string]int{"p1": 5, "p2": 3},
			lines: []Line{
				{ProductID: "p1", Quantity: 2},
				{ProductID: "p2", Quantity: 1},
			},
			orderID: "order-1",
			want: ReserveResult{Reserved: []Line{
				{ProductID: "p1", Quantity: 2},
				{ProductID: "p2", Quantity: 1},
			}},
			wantReserve: 1,
		},
		"insufficient item": {
			initialStock: map[string]int{"p1": 1, "p2": 0},
			lines: []Line{
				{ProductID: "p1", Quantity: 1},
				{ProductID: "p2", Quantity: 2},
			},
			orderID: "order-2",
			want: ReserveResult{
				Depleted: []DepletedLine{
					{ProductID: "p2", Requested: 2, Available: 0},
				},
			},
			wantReserve: 1,
		},
		"unknown product treated as depleted": {
			initialStock: map[string]int{"p1": 2},
			lines: []Line{
				{ProductID: "p1", Quantity: 1},
				{ProductID: "missing", Quantity: 1},
			},
			orderID: "order-3",
			want: ReserveResult{
				Depleted: []DepletedLine{
					{ProductID: "missing", Requested: 1, Available: 0},
				},
			},
			wantReserve: 1,
		},
		"repository error surfaces": {
			initialStock: map[string]int{"p1": 2},
			lines:        []Line{{ProductID: "p1", Quantity: 1}},
			orderID:      "order-4",
			repoErr:      errors.New("db down"),
			wantErr:      true,
			wantReserve:  1,
		},
		"idempotent reservation": {
			initialStock: map[string]int{"p1": 4},
			lines: []Line{
				{ProductID: "p1", Quantity: 3},
			},
			orderID: "order-5",
			want: ReserveResult{Reserved: []Line{
				{ProductID: "p1", Quantity: 3},
			}},
			wantReserve: 1,
		},
	}

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			repo := newFakeRepository(tt.initialStock)
			repo.reserveErr = tt.repoErr

			svc := NewService(repo)
			ctx := context.Background()

			res, err := svc.ReserveForOrder(ctx, tt.orderID, tt.lines)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(res, tt.want) {
				t.Fatalf("result mismatch\ngot  %+v\nwant %+v", res, tt.want)
			}

			// Trigger second call to verify idempotency behavior.
			res2, err := svc.ReserveForOrder(ctx, tt.orderID, tt.lines)
			if err != nil {
				t.Fatalf("second call error: %v", err)
			}
			if !reflect.DeepEqual(res2, res) {
				t.Fatalf("second call result mismatch\ngot  %+v\nwant %+v", res2, res)
			}

			if repo.reserveCnt != tt.wantReserve {
				t.Fatalf("Reserve called %d times, want %d", repo.reserveCnt, tt.wantReserve)
			}

			// For successful reservations, verify stock decreased only once.
			if len(tt.want.Reserved) > 0 {
				for _, ln := range tt.lines {
					if got := repo.stock[ln.ProductID]; got < 0 {
						t.Fatalf("negative stock for %s", ln.ProductID)
					}
				}
			}
		})
	}
}
