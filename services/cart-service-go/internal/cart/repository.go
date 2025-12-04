package cart

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Repository interface {
	GetCart(ctx context.Context, userID string) (*Cart, error)
	UpsertCart(ctx context.Context, c *Cart) error
	ClearCart(ctx context.Context, userID string) error
}

type repo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repo{db: db}
}

func (r *repo) GetCart(ctx context.Context, userID string) (*Cart, error) {
	const cartQuery = `SELECT id, user_id, total, updated_at FROM carts WHERE user_id = $1`

	var c Cart
	err := r.db.QueryRowContext(ctx, cartQuery, userID).Scan(&c.ID, &c.UserID, &c.Total, &c.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// caller (handler) can turn this into 404
			return nil, nil
		}
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT product_id, quantity, price FROM cart_items WHERE cart_id = $1`, c.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ProductID, &it.Quantity, &it.Price); err != nil {
			return nil, err
		}
		c.Items = append(c.Items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *repo) UpsertCart(ctx context.Context, c *Cart) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if c.ID == "" {
		c.ID = uuid.NewString()
	}

	const upsertCartSQL = `
INSERT INTO carts (id, user_id, total, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (user_id) DO UPDATE
SET total = EXCLUDED.total, updated_at = NOW()
RETURNING id, updated_at
`
	if err = tx.QueryRowContext(ctx, upsertCartSQL, c.ID, c.UserID, c.Total).Scan(&c.ID, &c.UpdatedAt); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, c.ID); err != nil {
		return err
	}

	if len(c.Items) > 0 {
		stmt, err := tx.PrepareContext(ctx, `INSERT INTO cart_items (id, cart_id, product_id, quantity, price) VALUES ($1, $2, $3, $4, $5)`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, it := range c.Items {
			if _, err = stmt.ExecContext(ctx, uuid.NewString(), c.ID, it.ProductID, it.Quantity, it.Price); err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	return err
}

func (r *repo) ClearCart(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM carts WHERE user_id = $1`, userID)
	return err
}
