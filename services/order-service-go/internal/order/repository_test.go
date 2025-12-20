package order

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestRepositoryCreate_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	now := time.Now()

	o := &Order{
		ID:          "order-123",
		CartID:      "cart-1",
		UserID:      "user-1",
		TotalAmount: 25.5,
		CreatedAt:   now,
		Items: []Item{
			{ProductID: "p1", Quantity: 1, Price: 10.0},
			{ProductID: "p2", Quantity: 2, Price: 7.75},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO orders (id, cart_id, user_id, total_amount, created_at)
         VALUES ($1, $2, $3, $4, $5)`)).
		WithArgs(o.ID, o.CartID, o.UserID, o.TotalAmount, o.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO order_items (id, order_id, product_id, quantity, price)
             VALUES ($1, $2, $3, $4, $5)`)).
		WithArgs(sqlmock.AnyArg(), o.ID, "p1", 1, 10.0).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO order_items (id, order_id, product_id, quantity, price)
             VALUES ($1, $2, $3, $4, $5)`)).
		WithArgs(sqlmock.AnyArg(), o.ID, "p2", 2, 7.75).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	require.NoError(t, repo.Create(ctx, o))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryCreate_OrderInsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	now := time.Now()

	o := &Order{
		ID:          "order-err",
		CartID:      "cart-err",
		UserID:      "user-err",
		TotalAmount: 10,
		CreatedAt:   now,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO orders (id, cart_id, user_id, total_amount, created_at)
         VALUES ($1, $2, $3, $4, $5)`)).
		WithArgs(o.ID, o.CartID, o.UserID, o.TotalAmount, o.CreatedAt).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	err = repo.Create(ctx, o)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryCreate_ItemInsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	now := time.Now()

	o := &Order{
		ID:          "order-item-err",
		CartID:      "cart-item",
		UserID:      "user-item",
		TotalAmount: 5,
		CreatedAt:   now,
		Items: []Item{
			{ProductID: "p1", Quantity: 1, Price: 5},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO orders (id, cart_id, user_id, total_amount, created_at)
         VALUES ($1, $2, $3, $4, $5)`)).
		WithArgs(o.ID, o.CartID, o.UserID, o.TotalAmount, o.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO order_items (id, order_id, product_id, quantity, price)
             VALUES ($1, $2, $3, $4, $5)`)).
		WithArgs(sqlmock.AnyArg(), o.ID, "p1", 1, 5.0).
		WillReturnError(errors.New("item insert failed"))
	mock.ExpectRollback()

	err = repo.Create(ctx, o)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryGetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, cart_id, user_id, total_amount, created_at
         FROM orders WHERE id = $1`)).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	o, err := repo.GetByID(context.Background(), "missing")
	require.NoError(t, err)
	require.Nil(t, o)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListByUser_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "cart_id", "user_id", "total_amount", "created_at",
		"product_id", "quantity", "price",
	})

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT
			o.id, o.cart_id, o.user_id, o.total_amount, o.created_at,
			oi.product_id, oi.quantity, oi.price
		FROM orders o
		LEFT JOIN order_items oi ON oi.order_id = o.id
		WHERE o.user_id = $1
		ORDER BY o.created_at DESC, o.id
	`)).
		WithArgs("user-empty").
		WillReturnRows(rows)

	orders, err := repo.ListByUser(context.Background(), "user-empty")
	require.NoError(t, err)
	require.Empty(t, orders)
	require.NoError(t, mock.ExpectationsWereMet())
}
