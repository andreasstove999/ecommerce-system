package integration

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"

	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/dedup"
	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/events"
	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/sequence"
	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/testutil"
)

func TestCartCheckedOutConsumer_PublishesOrderCreated(t *testing.T) {
	db, cleanupDB := testutil.StartPostgres(t)
	t.Cleanup(cleanupDB)

	repo := order.NewRepository(db)
	dedupRepo := dedup.NewRepository(db)
	seqRepo := sequence.NewRepository(db)

	conn, cleanupMQ := testutil.StartRabbitMQ(t)
	t.Cleanup(cleanupMQ)

	logger := log.New(io.Discard, "", 0)

	publisher, err := events.NewPublisher(conn, seqRepo, true)
	require.NoError(t, err)
	t.Cleanup(func() { _ = publisher.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	consumer := events.NewConsumer(conn, logger)
	consumer.Register(events.QueueCartCheckedOut, events.CartCheckedOutHandler(db, repo, dedupRepo, publisher, logger, true))

	require.NoError(t, consumer.Start(ctx))

	orderCreatedCh := make(chan events.OrderCreatedEnvelope, 1)

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

				var ev events.OrderCreatedEnvelope
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

	payload := events.CartCheckedOutPayload{
		CartID: "cart-200",
		UserID: "user-200",
		Items: []events.CartItem{
			{ProductID: "product-2", Quantity: 1, Price: 15.99},
		},
		TotalAmount: 15.99,
		Timestamp:   time.Now().UTC().Truncate(time.Millisecond),
	}
	sequenceVal := int64(1)
	correlationID := uuid.NewString()
	event := events.CartCheckedOutEnvelope{
		EventName:     "CartCheckedOut",
		EventVersion:  1,
		EventID:       uuid.NewString(),
		CorrelationID: correlationID,
		Producer:      "cart-service",
		PartitionKey:  payload.CartID,
		Sequence:      &sequenceVal,
		OccurredAt:    time.Now().UTC(),
		Schema:        "contracts/events/cart/CartCheckedOut.v1.payload.schema.json",
		Payload:       payload,
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

	var got events.OrderCreatedEnvelope
	require.Eventually(t, func() bool {
		select {
		case ev := <-orderCreatedCh:
			got = ev
			return true
		default:
			return false
		}
	}, 5*time.Second, 100*time.Millisecond)

	require.Equal(t, "OrderCreated", got.EventName)
	require.Equal(t, 1, got.EventVersion)
	require.Equal(t, correlationID, got.CorrelationID)
	require.Equal(t, event.EventID, got.CausationID)
	require.Equal(t, "order-service", got.Producer)
	require.Equal(t, got.Payload.OrderID, got.PartitionKey)
	require.NotNil(t, got.Sequence)
	require.Equal(t, int64(1), *got.Sequence)
	require.Equal(t, event.Payload.UserID, got.Payload.UserID)
	require.Equal(t, event.Payload.TotalAmount, got.Payload.TotalAmount)
}
