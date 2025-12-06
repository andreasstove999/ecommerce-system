package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, o *Order) error
	GetByID(ctx context.Context, orderID string) (*Order, error)
	ListByUser(ctx context.Context, userID string) ([]Order, error)
}

type repo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repo{db: db}
}

func (r *repo) Create(ctx context.Context, o *Order) error {
	if o.ID == "" {
		o.ID = uuid.NewString()
	}

	// TODO: Look into using transactions, and how they are handled in the application layer
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO orders (id, cart_id, user_id, total_amount, created_at)
         VALUES ($1, $2, $3, $4, $5)`,
		o.ID, o.CartID, o.UserID, o.TotalAmount, o.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	for _, it := range o.Items {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO order_items (id, order_id, product_id, quantity, price)
             VALUES ($1, $2, $3, $4, $5)`,
			uuid.NewString(), o.ID, it.ProductID, it.Quantity, it.Price,
		)
		if err != nil {
			return fmt.Errorf("insert order_item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (r *repo) GetByID(ctx context.Context, orderID string) (*Order, error) {
	var o Order
	err := r.db.QueryRowContext(ctx,
		`SELECT id, cart_id, user_id, total_amount, created_at
         FROM orders WHERE id = $1`,
		orderID,
	).Scan(&o.ID, &o.CartID, &o.UserID, &o.TotalAmount, &o.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("select order: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT product_id, quantity, price
         FROM order_items WHERE order_id = $1`,
		o.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("select order_items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ProductID, &it.Quantity, &it.Price); err != nil {
			return nil, fmt.Errorf("scan order_item: %w", err)
		}
		o.Items = append(o.Items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return &o, nil
}

func (r *repo) ListByUser(ctx context.Context, userID string) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, cart_id, user_id, total_amount, created_at
         FROM orders WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("select orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.CartID, &o.UserID, &o.TotalAmount, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	// Load items for each order - TODO: decide if this is the right way to do it, or if we should use a different approach for user or postman demo only
	for i := range orders {
		itemsRows, err := r.db.QueryContext(ctx,
			`SELECT product_id, quantity, price FROM order_items WHERE order_id = $1`,
			orders[i].ID,
		)
		if err != nil {
			return nil, fmt.Errorf("select items: %w", err)
		}
		for itemsRows.Next() {
			var it Item
			if err := itemsRows.Scan(&it.ProductID, &it.Quantity, &it.Price); err != nil {
				itemsRows.Close()
				return nil, fmt.Errorf("scan item: %w", err)
			}
			orders[i].Items = append(orders[i].Items, it)
		}
		itemsRows.Close()
	}

	return orders, nil
}
