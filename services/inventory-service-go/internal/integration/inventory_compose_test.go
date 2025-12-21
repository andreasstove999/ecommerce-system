//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/events"
	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/inventory"
	amqp "github.com/streadway/amqp"
)

const (
	inventoryBaseURL = "http://localhost:8083"
	rabbitURL        = "amqp://guest:guest@localhost:5672/"
)

func TestInventoryComposeFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := &http.Client{Timeout: 10 * time.Second}

	conn, err := amqp.Dial(rabbitURL)
	require.NoError(t, err)
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	purgeQueues(t, ch, events.StockReservedQueue, events.StockDepletedQueue)

	productA := "product-A"
	productB := "product-B"
	composeSeedStock(ctx, t, client, productA, 5)
	composeSeedStock(ctx, t, client, productB, 1)

	orderID1 := fmt.Sprintf("order-%d", time.Now().UnixNano())
	publishOrderCreatedMessage(ctx, t, ch, events.OrderCreated{
		EventType: events.EventTypeOrderCreated,
		OrderID:   orderID1,
		UserID:    "user-1",
		Timestamp: time.Now().UTC(),
		Items: []events.CartItem{
			{ProductID: productA, Quantity: 2},
		},
	})

	reserved := waitForQueueMessage[events.StockReserved](ctx, t, ch, events.StockReservedQueue)
	require.Equal(t, orderID1, reserved.OrderID)
	require.Len(t, reserved.Items, 1)
	require.Equal(t, productA, reserved.Items[0].ProductID)
	require.Equal(t, 2, reserved.Items[0].Quantity)

	waitForStockAvailability(ctx, t, client, productA, 3)
	waitForStockAvailability(ctx, t, client, productB, 1)

	orderID2 := uuid.NewString()
	publishOrderCreatedMessage(ctx, t, ch, events.OrderCreated{
		EventType: events.EventTypeOrderCreated,
		OrderID:   orderID2,
		UserID:    "user-2",
		Timestamp: time.Now().UTC(),
		Items: []events.CartItem{
			{ProductID: productA, Quantity: 2},
			{ProductID: productB, Quantity: 2},
		},
	})

	depleted := waitForQueueMessage[events.StockDepleted](ctx, t, ch, events.StockDepletedQueue)
	require.Equal(t, orderID2, depleted.OrderID)
	require.Len(t, depleted.Depleted, 1)
	require.Equal(t, productB, depleted.Depleted[0].ProductID)
	require.Equal(t, 2, depleted.Depleted[0].Requested)
	require.Equal(t, 1, depleted.Depleted[0].Available)

	waitForStockAvailability(ctx, t, client, productA, 3)
	waitForStockAvailability(ctx, t, client, productB, 1)
}

func composeSeedStock(ctx context.Context, t *testing.T, client *http.Client, productID string, available int) {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"productId": productID,
		"available": available,
	})
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/inventory/adjust", inventoryBaseURL), bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func publishOrderCreatedMessage(ctx context.Context, t *testing.T, ch *amqp.Channel, order events.OrderCreated) {
	t.Helper()

	_, err := ch.QueueDeclare(events.QueueOrderCreated, true, false, false, false, nil)
	require.NoError(t, err)

	body, err := json.Marshal(order)
	require.NoError(t, err)

	pubCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	require.NoError(t, ch.PublishWithContext(pubCtx, "", events.QueueOrderCreated, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	}))
}

func waitForQueueMessage[T any](ctx context.Context, t *testing.T, ch *amqp.Channel, queue string) T {
	t.Helper()

	_, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	require.NoError(t, err)

	var out T
	pollCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	backoff := 50 * time.Millisecond
	for {
		select {
		case <-pollCtx.Done():
			t.Fatalf("timed out waiting for message on %s: %v", queue, pollCtx.Err())
		default:
		}

		msg, ok, getErr := ch.Get(queue, false)
		require.NoError(t, getErr)
		if ok {
			require.NoError(t, json.Unmarshal(msg.Body, &out))
			require.NoError(t, msg.Ack(false))
			return out
		}

		time.Sleep(backoff)
		if backoff < time.Second {
			backoff *= 2
			if backoff > time.Second {
				backoff = time.Second
			}
		}
	}
}

func waitForStockAvailability(ctx context.Context, t *testing.T, client *http.Client, productID string, expected int) inventory.StockItem {
	t.Helper()

	var item inventory.StockItem
	pollCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	backoff := 50 * time.Millisecond
	for {
		select {
		case <-pollCtx.Done():
			t.Fatalf("timed out waiting for availability for %s: %v", productID, pollCtx.Err())
		default:
		}

		req, err := http.NewRequestWithContext(pollCtx, http.MethodGet, fmt.Sprintf("%s/api/inventory/%s", inventoryBaseURL, productID), nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)

		if resp.StatusCode == http.StatusOK {
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&item))
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK && item.Available == expected {
			return item
		}

		time.Sleep(backoff)
		if backoff < time.Second {
			backoff *= 2
			if backoff > time.Second {
				backoff = time.Second
			}
		}
	}
}

func purgeQueues(t *testing.T, ch *amqp.Channel, queues ...string) {
	t.Helper()
	for _, q := range queues {
		_, err := ch.QueueDeclare(q, true, false, false, false, nil)
		require.NoError(t, err)
		_, err = ch.QueuePurge(q, false)
		require.NoError(t, err)
	}
}
