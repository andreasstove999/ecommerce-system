package inventory

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestPostgresRepository_Get(t *testing.T) {
	ctx := context.Background()
	pool := newMockPool(map[string]int{"p1": 7})
	repo := NewPostgresRepository(pool)

	item, err := repo.Get(ctx, "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ProductID != "p1" || item.Available != 7 {
		t.Fatalf("unexpected item: %+v", item)
	}
}

func TestPostgresRepository_GetMissing(t *testing.T) {
	ctx := context.Background()
	repo := NewPostgresRepository(newMockPool(nil))

	_, err := repo.Get(ctx, "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestPostgresRepository_SetAvailable(t *testing.T) {
	ctx := context.Background()
	pool := newMockPool(nil)
	repo := NewPostgresRepository(pool)

	if err := repo.SetAvailable(ctx, "p1", 10); err != nil {
		t.Fatalf("set available: %v", err)
	}
	if err := repo.SetAvailable(ctx, "p1", 4); err != nil {
		t.Fatalf("update available: %v", err)
	}

	if got := pool.stocks["p1"]; got != 4 {
		t.Fatalf("stock not updated, got %d", got)
	}
}

func TestPostgresRepository_Reserve(t *testing.T) {
	ctx := context.Background()

	t.Run("reserves atomically", func(t *testing.T) {
		pool := newMockPool(map[string]int{"p1": 5, "p2": 3})
		repo := NewPostgresRepository(pool)

		result, err := repo.Reserve(ctx, "order-1", []Line{
			{ProductID: "p1", Quantity: 2},
			{ProductID: "p2", Quantity: 1},
		})
		if err != nil {
			t.Fatalf("reserve: %v", err)
		}

		want := ReserveResult{
			Reserved: []Line{
				{ProductID: "p1", Quantity: 2},
				{ProductID: "p2", Quantity: 1},
			},
		}
		if !reflect.DeepEqual(result, want) {
			t.Fatalf("result mismatch\ngot  %+v\nwant %+v", result, want)
		}
		if pool.stocks["p1"] != 3 || pool.stocks["p2"] != 2 {
			t.Fatalf("stocks not decremented: %+v", pool.stocks)
		}
		if pool.lastTx == nil || !pool.lastTx.committed || pool.lastTx.rolledBack {
			t.Fatalf("transaction state incorrect: %+v", pool.lastTx)
		}
	})

	t.Run("insufficient stock rolls back", func(t *testing.T) {
		pool := newMockPool(map[string]int{"p1": 1})
		repo := NewPostgresRepository(pool)

		result, err := repo.Reserve(ctx, "order-2", []Line{
			{ProductID: "p1", Quantity: 2},
		})
		if err != nil {
			t.Fatalf("reserve: %v", err)
		}
		if len(result.Depleted) != 1 || result.Depleted[0].ProductID != "p1" {
			t.Fatalf("unexpected depleted: %+v", result.Depleted)
		}
		if pool.stocks["p1"] != 1 {
			t.Fatalf("stock mutated despite depletion: %d", pool.stocks["p1"])
		}
	})

	t.Run("unknown product treated as zero available", func(t *testing.T) {
		pool := newMockPool(map[string]int{"p1": 2})
		repo := NewPostgresRepository(pool)

		result, err := repo.Reserve(ctx, "order-3", []Line{
			{ProductID: "missing", Quantity: 1},
		})
		if err != nil {
			t.Fatalf("reserve: %v", err)
		}
		if len(result.Depleted) != 1 || result.Depleted[0].Available != 0 {
			t.Fatalf("expected depleted with zero availability, got %+v", result.Depleted)
		}
		if pool.stocks["p1"] != 2 {
			t.Fatalf("stock should be unchanged: %d", pool.stocks["p1"])
		}
	})

	t.Run("begin transaction error surfaces", func(t *testing.T) {
		pool := newMockPool(nil)
		pool.beginErr = errors.New("cannot begin")
		repo := NewPostgresRepository(pool)

		if _, err := repo.Reserve(ctx, "order-4", []Line{{ProductID: "p1", Quantity: 1}}); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("exec failure rolls back without applying changes", func(t *testing.T) {
		pool := newMockPool(map[string]int{"p1": 3})
		pool.execErr = errors.New("update fail")
		repo := NewPostgresRepository(pool)

		if _, err := repo.Reserve(ctx, "order-5", []Line{{ProductID: "p1", Quantity: 1}}); err == nil {
			t.Fatalf("expected exec error")
		}
		if pool.stocks["p1"] != 3 {
			t.Fatalf("stock changed after exec error: %d", pool.stocks["p1"])
		}
		if pool.lastTx == nil || !pool.lastTx.rolledBack {
			t.Fatalf("transaction not rolled back after exec failure")
		}
	})

	t.Run("commit failure does not persist updates", func(t *testing.T) {
		pool := newMockPool(map[string]int{"p1": 2})
		repo := NewPostgresRepository(pool)
		pool.commitErr = errors.New("commit fail")

		if _, err := repo.Reserve(ctx, "order-6", []Line{{ProductID: "p1", Quantity: 1}}); err == nil {
			t.Fatalf("expected commit error")
		}
		if pool.stocks["p1"] != 2 {
			t.Fatalf("stock changed after commit failure: %d", pool.stocks["p1"])
		}
		if pool.lastTx == nil || !pool.lastTx.rolledBack {
			t.Fatalf("rollback not invoked after commit failure")
		}
	})
}

type mockPool struct {
	stocks map[string]int

	beginErr  error
	execErr   error
	commitErr error

	lastTx  *mockTx
	txCount int
}

func newMockPool(initial map[string]int) *mockPool {
	cp := make(map[string]int, len(initial))
	for k, v := range initial {
		cp[k] = v
	}
	return &mockPool{stocks: cp}
}

func (p *mockPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	productID := args[0].(string)
	available, ok := p.stocks[productID]
	if !ok {
		return mockRow{err: pgx.ErrNoRows}
	}

	if strings.Contains(sql, "SELECT product_id") {
		return mockRow{values: []any{productID, available}}
	}
	return mockRow{values: []any{available}}
}

func (p *mockPool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	productID := args[0].(string)
	available := args[1].(int)
	p.stocks[productID] = available
	return pgconn.CommandTag("EXEC"), p.execErr
}

func (p *mockPool) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgxTx, error) {
	if p.beginErr != nil {
		return nil, p.beginErr
	}
	p.txCount++
	tx := &mockTx{
		pool:    p,
		pending: make(map[string]int),
		execErr: p.execErr,
	}
	p.lastTx = tx
	return tx, nil
}

type mockTx struct {
	pool *mockPool

	pending map[string]int

	execErr   error
	commitErr error

	rolledBack bool
	committed  bool
}

func (tx *mockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	productID := args[0].(string)
	available, ok := tx.pool.stocks[productID]
	if !ok {
		return mockRow{err: pgx.ErrNoRows}
	}
	return mockRow{values: []any{available}}
}

func (tx *mockTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if tx.execErr != nil {
		return pgconn.CommandTag(""), tx.execErr
	}
	productID := args[0].(string)
	quantity := args[1].(int)
	tx.pending[productID] += quantity
	return pgconn.CommandTag("EXEC"), nil
}

func (tx *mockTx) Commit(ctx context.Context) error {
	if tx.commitErr != nil || tx.pool.commitErr != nil {
		return errors.Join(tx.commitErr, tx.pool.commitErr)
	}
	for productID, dec := range tx.pending {
		tx.pool.stocks[productID] -= dec
	}
	tx.committed = true
	return nil
}

func (tx *mockTx) Rollback(ctx context.Context) error {
	tx.rolledBack = true
	return nil
}

type mockRow struct {
	values []any
	err    error
}

func (r mockRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	for i, v := range r.values {
		switch d := dest[i].(type) {
		case *string:
			*d = v.(string)
		case *int:
			*d = v.(int)
		default:
			return errors.New("unsupported scan dest")
		}
	}
	return nil
}
