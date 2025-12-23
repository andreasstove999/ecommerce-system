package inventory

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrNotFound = errors.New("not found")

// DBPool matches the methods from *pgxpool.Pool that we use.
// This allows us to mock the database in tests.
type DBPool interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type Repository interface {
	Get(ctx context.Context, productID string) (StockItem, error)
	SetAvailable(ctx context.Context, productID string, available int) error
	Reserve(ctx context.Context, orderID string, lines []Line) (ReserveResult, error)
}

type TransactionalRepository interface {
	Repository
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	ReserveWithTx(ctx context.Context, tx pgx.Tx, orderID string, lines []Line) (ReserveResult, error)
}

type PostgresRepository struct {
	pool DBPool
}

func NewPostgresRepository(pool DBPool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Get(ctx context.Context, productID string) (StockItem, error) {
	var item StockItem
	row := r.pool.QueryRow(ctx, `SELECT product_id, available FROM inventory_stock WHERE product_id=$1`, productID)
	if err := row.Scan(&item.ProductID, &item.Available); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StockItem{}, ErrNotFound
		}
		return StockItem{}, err
	}
	return item, nil
}

func (r *PostgresRepository) SetAvailable(ctx context.Context, productID string, available int) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO inventory_stock(product_id, available)
		VALUES($1, $2)
		ON CONFLICT (product_id) DO UPDATE SET available=EXCLUDED.available, updated_at=now()
	`, productID, available)
	return err
}

func (r *PostgresRepository) Reserve(ctx context.Context, orderID string, lines []Line) (ReserveResult, error) {
	// This is a minimal “atomic reserve” implementation:
	// - locks each product row (SELECT ... FOR UPDATE)
	// - if any line is short, we rollback and return depleted info (no mutation)
	// - else we decrement stock for all lines and commit
	//
	// NOTE: There is no idempotency/reservation table here yet.
	// If you need exactly-once semantics, add an inventory_reservations table keyed by order_id.

	res := ReserveResult{}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return res, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	res, err = r.reserveWithTx(ctx, tx, orderID, lines)
	if err != nil {
		return res, err
	}

	if len(res.Depleted) > 0 {
		return res, nil
	}

	if err := tx.Commit(ctx); err != nil {
		return res, err
	}
	return res, nil
}

func (r *PostgresRepository) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return r.pool.BeginTx(ctx, txOptions)
}

func (r *PostgresRepository) ReserveWithTx(ctx context.Context, tx pgx.Tx, orderID string, lines []Line) (ReserveResult, error) {
	return r.reserveWithTx(ctx, tx, orderID, lines)
}

func (r *PostgresRepository) reserveWithTx(ctx context.Context, tx pgx.Tx, orderID string, lines []Line) (ReserveResult, error) {
	res := ReserveResult{}

	type locked struct {
		productID string
		requested int
		available int
	}
	lockedRows := make([]locked, 0, len(lines))

	for _, line := range lines {
		var available int
		err := tx.QueryRow(ctx, `
			SELECT available
			FROM inventory_stock
			WHERE product_id=$1
			FOR UPDATE
		`, line.ProductID).Scan(&available)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				available = 0
			} else {
				return res, err
			}
		}

		lockedRows = append(lockedRows, locked{productID: line.ProductID, requested: line.Quantity, available: available})
		if available < line.Quantity {
			res.Depleted = append(res.Depleted, DepletedLine{
				ProductID: line.ProductID,
				Requested: line.Quantity,
				Available: available,
			})
		}
	}

	if len(res.Depleted) > 0 {
		return res, nil
	}

	for _, row := range lockedRows {
		_, err := tx.Exec(ctx, `
			UPDATE inventory_stock
			SET available = available - $2, updated_at=now()
			WHERE product_id=$1
		`, row.productID, row.requested)
		if err != nil {
			return res, err
		}
		res.Reserved = append(res.Reserved, Line{ProductID: row.productID, Quantity: row.requested})
	}

	return res, nil
}
