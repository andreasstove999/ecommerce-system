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

func TestCartCheckedOutConsumer_PublishesOrderCreated(t *testing.T) {
	db, cleanupDB := testutil.StartPostgres(t)
	t.Cleanup(cleanupDB)

	repo := order.NewRepository(db)

	conn, cleanupMQ := testutil.StartRabbitMQ(t)
	t.Cleanup(cleanupMQ)

	logger := log.New(io.Discard, "", 0)

	publisher, err := events.NewPublisher(conn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = publisher.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	consumer := events.NewConsumer(conn, logger)
	consumer.Register(events.QueueCartCheckedOut, events.CartCheckedOutHandler(repo, publisher, logger))

	require.NoError(t, consumer.Start(ctx))

	orderCreatedCh := make(chan events.OrderCreated, 1)

	consumeCh, err := conn.Channel()
	require.NoError(t, err)
	t.Cleanup(func() { _ = consumeCh.Close() })

	_, err = consumeCh.QueueDeclare(
		events.OrderCreatedQueue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	require.NoError(t, err)

	msgs, err := consumeCh.Consume(
		events.OrderCreatedQueue,
		"integration-order-created",
		true,  // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	require.NoError(t, err)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				var ev events.OrderCreated
				if err := json.Unmarshal(msg.Body, &ev); err != nil {
					continue
				}

				orderCreatedCh <- ev
				return
			}
		}
	}()

	publishCh, err := conn.Channel()
	require.NoError(t, err)
	t.Cleanup(func() { _ = publishCh.Close() })

	_, err = publishCh.QueueDeclare(
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
		CartID:    "cart-200",
		UserID:    "user-200",
		Items: []events.CartItem{
			{ProductID: "product-2", Quantity: 1, Price: 15.99},
		},
		TotalAmount: 15.99,
		Timestamp:   time.Now().UTC().Truncate(time.Millisecond),
	}

	body, err := json.Marshal(event)
	require.NoError(t, err)

	err = publishCh.PublishWithContext(
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

	var got events.OrderCreated
	require.Eventually(t, func() bool {
		select {
		case ev := <-orderCreatedCh:
			got = ev
			return true
		default:
			return false
		}
	}, 5*time.Second, 100*time.Millisecond)

	require.Equal(t, "OrderCreated", got.EventType)
	require.Equal(t, event.UserID, got.UserID)
	require.Equal(t, event.TotalAmount, got.TotalAmount)
}
