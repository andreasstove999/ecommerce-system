package inventory

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresRepository(mock)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT product_id, available FROM inventory_stock WHERE product_id=$1")).
		WithArgs("prod-1").
		WillReturnRows(mock.NewRows([]string{"product_id", "available"}).AddRow("prod-1", 10))

	item, err := repo.Get(context.Background(), "prod-1")
	require.NoError(t, err)
	assert.Equal(t, "prod-1", item.ProductID)
	assert.Equal(t, 10, item.Available)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGet_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresRepository(mock)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT product_id, available FROM inventory_stock WHERE product_id=$1")).
		WithArgs("prod-missing").
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.Get(context.Background(), "prod-missing")
	require.ErrorIs(t, err, ErrNotFound)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSetAvailable_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresRepository(mock)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO inventory_stock(product_id, available)")).
		WithArgs("prod-1", 100).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.SetAvailable(context.Background(), "prod-1", 100)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReserve_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresRepository(mock)

	mock.ExpectBeginTx(pgx.TxOptions{})

	// Check item 1
	mock.ExpectQuery(regexp.QuoteMeta("SELECT available FROM inventory_stock WHERE product_id=$1 FOR UPDATE")).
		WithArgs("p1").
		WillReturnRows(mock.NewRows([]string{"available"}).AddRow(10))

	// Check item 2
	mock.ExpectQuery(regexp.QuoteMeta("SELECT available FROM inventory_stock WHERE product_id=$1 FOR UPDATE")).
		WithArgs("p2").
		WillReturnRows(mock.NewRows([]string{"available"}).AddRow(5))

	// Update item 1
	mock.ExpectExec(regexp.QuoteMeta("UPDATE inventory_stock SET available = available - $2, updated_at=now() WHERE product_id=$1")).
		WithArgs("p1", 2).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	// Update item 2
	mock.ExpectExec(regexp.QuoteMeta("UPDATE inventory_stock SET available = available - $2, updated_at=now() WHERE product_id=$1")).
		WithArgs("p2", 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectCommit()

	res, err := repo.Reserve(context.Background(), "order-1", []Line{
		{ProductID: "p1", Quantity: 2},
		{ProductID: "p2", Quantity: 1},
	})
	require.NoError(t, err)
	assert.Empty(t, res.Depleted)
	assert.Len(t, res.Reserved, 2)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReserve_Depleted(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresRepository(mock)

	// Explicitly expect Rollback because the repository defers a rollback.
	// Even though we don't return error, the code path is:
	// Begin -> ... -> return early -> defer rollback
	mock.ExpectBeginTx(pgx.TxOptions{})

	// Check item 1 - OK
	mock.ExpectQuery(regexp.QuoteMeta("SELECT available FROM inventory_stock WHERE product_id=$1 FOR UPDATE")).
		WithArgs("p1").
		WillReturnRows(mock.NewRows([]string{"available"}).AddRow(10))

	// Check item 2 - Not enough
	mock.ExpectQuery(regexp.QuoteMeta("SELECT available FROM inventory_stock WHERE product_id=$1 FOR UPDATE")).
		WithArgs("p2").
		WillReturnRows(mock.NewRows([]string{"available"}).AddRow(2)) // Requesting 5

	mock.ExpectRollback()

	res, err := repo.Reserve(context.Background(), "order-1", []Line{
		{ProductID: "p1", Quantity: 2},
		{ProductID: "p2", Quantity: 5},
	})
	require.NoError(t, err)
	assert.Len(t, res.Depleted, 1)
	assert.Equal(t, "p2", res.Depleted[0].ProductID)
	assert.Empty(t, res.Reserved)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReserve_DBError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresRepository(mock)

	mock.ExpectBeginTx(pgx.TxOptions{})
	mock.ExpectQuery(regexp.QuoteMeta("SELECT available FROM inventory_stock WHERE product_id=$1 FOR UPDATE")).
		WithArgs("p1").
		WillReturnError(errors.New("db boom"))
	mock.ExpectRollback()

	_, err = repo.Reserve(context.Background(), "order-1", []Line{{ProductID: "p1", Quantity: 1}})
	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}
