package integration

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"

	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/events"
	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/testutil"
)

func TestCartCheckedOutConsumer_CreatesOrder(t *testing.T) {
	db, cleanupDB := testutil.StartPostgres(t)
	t.Cleanup(cleanupDB)

	repo := order.NewRepository(db)

	conn, cleanupMQ := testutil.StartRabbitMQ(t)
	t.Cleanup(cleanupMQ)

	logger := log.New(io.Discard, "", 0)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, closePublisher, err := events.StartCartCheckedOutConsumer(ctx, conn, repo, logger)
	require.NoError(t, err)
	if closePublisher != nil {
		t.Cleanup(closePublisher)
	}

	ch, err := conn.Channel()
	require.NoError(t, err)
	t.Cleanup(func() { _ = ch.Close() })

	_, err = ch.QueueDeclare(
		events.QueueCartCheckedOut,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	require.NoError(t, err)

	event := events.CartCheckedOut{
		EventType: "CartCheckedOut",
		CartID:    "cart-100",
		UserID:    "user-100",
		Items: []events.CartItem{
			{ProductID: "product-1", Quantity: 2, Price: 9.99},
		},
		TotalAmount: 19.98,
		Timestamp:   time.Now().UTC().Truncate(time.Millisecond),
	}

	body, err := json.Marshal(event)
	require.NoError(t, err)

	err = ch.PublishWithContext(
		ctx,
		"",
		events.QueueCartCheckedOut,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		orders, err := repo.ListByUser(ctx, event.UserID)
		if err != nil {
			return false
		}
		if len(orders) != 1 {
			return false
		}
		return len(orders[0].Items) > 0
	}, 5*time.Second, 100*time.Millisecond)

	orders, err := repo.ListByUser(ctx, event.UserID)
	require.NoError(t, err)
	require.Len(t, orders, 1)
	require.NotEmpty(t, orders[0].Items)
}
