package integration

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/testutil"
)

func TestRepository_CreateAndGetByID(t *testing.T) {
	db, cleanup := testutil.StartPostgres(t)
	t.Cleanup(cleanup)
	truncateTables(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	repo := order.NewRepository(db)

	createdAt := time.Now().UTC().Truncate(time.Millisecond)
	orderToCreate := order.Order{
		CartID:      "cart-123",
		UserID:      "user-abc",
		TotalAmount: 42.50,
		CreatedAt:   createdAt,
		Items: []order.Item{
			{ProductID: "product-1", Quantity: 1, Price: 10.00},
			{ProductID: "product-2", Quantity: 3, Price: 32.50},
		},
	}

	require.NoError(t, repo.Create(ctx, &orderToCreate))

	fetched, err := repo.GetByID(ctx, orderToCreate.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, orderToCreate.ID, fetched.ID)
	require.Equal(t, orderToCreate.CartID, fetched.CartID)
	require.Equal(t, orderToCreate.UserID, fetched.UserID)
	require.Equal(t, orderToCreate.TotalAmount, fetched.TotalAmount)
	require.WithinDuration(t, orderToCreate.CreatedAt, fetched.CreatedAt, time.Millisecond)
	require.Len(t, fetched.Items, 2)
	require.ElementsMatch(t, orderToCreate.Items, fetched.Items)
}

func TestRepository_ListByUser_ReturnsOrdersWithItems(t *testing.T) {
	db, cleanup := testutil.StartPostgres(t)
	t.Cleanup(cleanup)
	truncateTables(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	repo := order.NewRepository(db)

	userID := "user-list"
	now := time.Now().UTC().Truncate(time.Millisecond)

	olderOrder := order.Order{
		CartID:      "cart-old",
		UserID:      userID,
		TotalAmount: 15.00,
		CreatedAt:   now.Add(-10 * time.Minute),
		Items: []order.Item{
			{ProductID: "product-old", Quantity: 2, Price: 7.50},
		},
	}

	newerOrder := order.Order{
		CartID:      "cart-new",
		UserID:      userID,
		TotalAmount: 30.00,
		CreatedAt:   now,
		Items: []order.Item{
			{ProductID: "product-new", Quantity: 1, Price: 30.00},
		},
	}

	require.NoError(t, repo.Create(ctx, &olderOrder))
	require.NoError(t, repo.Create(ctx, &newerOrder))

	orders, err := repo.ListByUser(ctx, userID)
	require.NoError(t, err)

	require.Len(t, orders, 2)
	require.Equal(t, newerOrder.ID, orders[0].ID)
	require.Equal(t, olderOrder.ID, orders[1].ID)
	require.True(t, orders[0].CreatedAt.After(orders[1].CreatedAt))

	require.Len(t, orders[0].Items, 1)
	require.Equal(t, newerOrder.Items[0], orders[0].Items[0])
	require.Len(t, orders[1].Items, 1)
	require.Equal(t, olderOrder.Items[0], orders[1].Items[0])
}

func truncateTables(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, `TRUNCATE order_items, orders, event_sequence, event_dedup_checkpoint`)
	require.NoError(t, err)
}
