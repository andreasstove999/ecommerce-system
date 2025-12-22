package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, o *Order) error
	CreateWithTx(ctx context.Context, tx *sql.Tx, o *Order) error
	GetByID(ctx context.Context, orderID string) (*Order, error)
	ListByUser(ctx context.Context, userID string) ([]Order, error)
	MarkPaymentSucceeded(ctx context.Context, orderID string) (*CompletionState, error)
	MarkPaymentFailed(ctx context.Context, orderID string, reason string) error
	MarkStockReserved(ctx context.Context, orderID string) (*CompletionState, error)
	MarkCompleted(ctx context.Context, orderID string) error
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

func (r *repo) CreateWithTx(ctx context.Context, tx *sql.Tx, o *Order) error {
	if o.ID == "" {
		o.ID = uuid.NewString()
	}

	_, err := tx.ExecContext(ctx,
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
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			o.id, o.cart_id, o.user_id, o.total_amount, o.created_at,
			oi.product_id, oi.quantity, oi.price
		FROM orders o
		LEFT JOIN order_items oi ON oi.order_id = o.id
		WHERE o.user_id = $1
		ORDER BY o.created_at DESC, o.id
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("select orders+items: %w", err)
	}
	defer rows.Close()

	orders := make([]Order, 0)
	indexByID := make(map[string]int)

	for rows.Next() {
		var (
			orderID     string
			cartID      string
			uID         string
			totalAmount float64
			createdAt   time.Time

			// LEFT JOIN: item columns may be NULL
			productID sql.NullString
			qty       sql.NullInt64
			price     sql.NullFloat64
		)

		if err := rows.Scan(
			&orderID, &cartID, &uID, &totalAmount, &createdAt,
			&productID, &qty, &price,
		); err != nil {
			return nil, fmt.Errorf("scan orders+items: %w", err)
		}

		idx, exists := indexByID[orderID]
		if !exists {
			orders = append(orders, Order{
				ID:          orderID,
				CartID:      cartID,
				UserID:      uID,
				TotalAmount: totalAmount,
				CreatedAt:   createdAt,
				Items:       []Item{},
			})
			idx = len(orders) - 1
			indexByID[orderID] = idx
		}

		// If there is an item row, append it
		if productID.Valid {
			orders[idx].Items = append(orders[idx].Items, Item{
				ProductID: productID.String,
				Quantity:  int(qty.Int64),
				Price:     price.Float64,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return orders, nil
}

type CompletionState struct {
	UserID          string
	ReadyToComplete bool
}

func (r *repo) MarkPaymentSucceeded(ctx context.Context, orderID string) (*CompletionState, error) {
	_, err := r.db.ExecContext(ctx,
		`UPDATE orders
		 SET payment_ok = true
		 WHERE id = $1`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("update payment_ok: %w", err)
	}

	return r.completionState(ctx, orderID)
}

func (r *repo) MarkPaymentFailed(ctx context.Context, orderID string, reason string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE orders
		 SET status = 'payment_failed',
		     payment_ok = false,
		     payment_error = $2
		 WHERE id = $1`,
		orderID, reason,
	)
	if err != nil {
		return fmt.Errorf("update payment_failed: %w", err)
	}
	return nil
}

func (r *repo) MarkStockReserved(ctx context.Context, orderID string) (*CompletionState, error) {
	_, err := r.db.ExecContext(ctx,
		`UPDATE orders
		 SET stock_ok = true
		 WHERE id = $1`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("update stock_ok: %w", err)
	}
	return r.completionState(ctx, orderID)
}

func (r *repo) MarkCompleted(ctx context.Context, orderID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status = 'completed' WHERE id = $1`,
		orderID,
	)
	if err != nil {
		return fmt.Errorf("update status completed: %w", err)
	}
	return nil
}

func (r *repo) completionState(ctx context.Context, orderID string) (*CompletionState, error) {
	var (
		userID    string
		paymentOK bool
		stockOK   bool
		status    string
	)

	err := r.db.QueryRowContext(ctx,
		`SELECT user_id, payment_ok, stock_ok, status
		 FROM orders WHERE id = $1`,
		orderID,
	).Scan(&userID, &paymentOK, &stockOK, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select completion state: %w", err)
	}

	// Do not complete failed/cancelled orders
	if status == "payment_failed" || status == "cancelled" {
		return &CompletionState{UserID: userID, ReadyToComplete: false}, nil
	}

	return &CompletionState{
		UserID:          userID,
		ReadyToComplete: paymentOK && stockOK,
	}, nil
}
