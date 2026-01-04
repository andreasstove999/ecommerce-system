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

func TestCartCheckedOutConsumer_CreatesOrder(t *testing.T) {
	db, cleanupDB := testutil.StartPostgres(t)
	t.Cleanup(cleanupDB)

	repo := order.NewRepository(db)
	dedupRepo := dedup.NewRepository(db)
	seqRepo := sequence.NewRepository(db)

	conn, cleanupMQ := testutil.StartRabbitMQ(t)
	t.Cleanup(cleanupMQ)

	logger := log.New(io.Discard, "", 0)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, closePublisher, err := events.StartCartCheckedOutConsumer(ctx, conn, db, repo, dedupRepo, seqRepo, logger, true, true)
	require.NoError(t, err)
	if closePublisher != nil {
		t.Cleanup(closePublisher)
	}

	ch, err := conn.Channel()
	require.NoError(t, err)
	t.Cleanup(func() { _ = ch.Close() })

	require.NoError(t, ch.ExchangeDeclare(events.EventsExchange, "topic", true, false, false, false, nil))

	payload := events.CartCheckedOutPayload{
		CartID: "cart-100",
		UserID: "user-100",
		Items: []events.CartItem{
			{ProductID: "product-1", Quantity: 2, Price: 9.99},
		},
		TotalAmount: 19.98,
		Timestamp:   time.Now().UTC().Truncate(time.Millisecond),
	}

	sequenceVal := int64(1)
	event := events.CartCheckedOutEnvelope{
		EventName:     "CartCheckedOut",
		EventVersion:  1,
		EventID:       uuid.NewString(),
		CorrelationID: uuid.NewString(),
		Producer:      "cart-service",
		PartitionKey:  payload.CartID,
		Sequence:      &sequenceVal,
		OccurredAt:    time.Now().UTC(),
		Schema:        "contracts/events/cart/CartCheckedOut.v1.payload.schema.json",
		Payload:       payload,
	}

	body, err := json.Marshal(event)
	require.NoError(t, err)

	err = ch.PublishWithContext(
		ctx,
		events.EventsExchange,
		events.CartCheckedOutRoutingKey,
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
		orders, err := repo.ListByUser(ctx, event.Payload.UserID)
		if err != nil {
			return false
		}
		if len(orders) != 1 {
			return false
		}
		return len(orders[0].Items) > 0
	}, 5*time.Second, 100*time.Millisecond)

	orders, err := repo.ListByUser(ctx, event.Payload.UserID)
	require.NoError(t, err)
	require.Len(t, orders, 1)
	require.NotEmpty(t, orders[0].Items)
}
