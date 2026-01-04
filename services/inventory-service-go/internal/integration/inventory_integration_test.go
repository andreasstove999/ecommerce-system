package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/db"
	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/events"
	httpapi "github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/http"
	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/inventory"
)

const (
	productA = "product-A"
	productB = "product-B"
)

func TestInventoryIntegration(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pgC, dbURL := startPostgres(ctx, t)
	defer terminateContainer(t, pgC)

	rabbitC, rabbitURL := startRabbitMQ(ctx, t)
	defer terminateContainer(t, rabbitC)

	logger := log.New(io.Discard, "", log.LstdFlags)
	require.NoError(t, db.RunMigrations(dbURL, logger))

	app := startInventoryService(ctx, t, dbURL, rabbitURL)
	defer app.stop()

	client := &http.Client{Timeout: 5 * time.Second}
	seedStock(ctx, t, client, app.baseURL, productA, 5)
	seedStock(ctx, t, client, app.baseURL, productB, 1)

	orderConn := dialAMQP(ctx, t, rabbitURL)
	defer orderConn.Close()

	orderWithStockID := "f1e2d3c4-b5a6-4988-99aa-bbccddeeff11"
	orderWithStock := newOrderCreated(orderWithStockID, "user-1", []events.OrderLineItem{
		{ProductID: productA, Quantity: 2},
	}, 1)
	publishOrderCreated(ctx, t, orderConn, orderWithStock)
	reserved := waitForStockReserved(ctx, t, orderConn)
	require.Equal(t, orderWithStock.Payload.OrderID, reserved.Payload.OrderID)
	require.Equal(t, reserved.EventEnvelope.PartitionKey, reserved.Payload.OrderID)
	require.Len(t, reserved.Payload.Items, 1)
	require.Equal(t, productA, reserved.Payload.Items[0].ProductID)
	require.Equal(t, 2, reserved.Payload.Items[0].Quantity)
	require.Equal(t, orderWithStock.EventEnvelope.CorrelationID, reserved.EventEnvelope.CorrelationID)
	require.Equal(t, orderWithStock.EventEnvelope.EventID, reserved.EventEnvelope.CausationID)

	waitForAvailability(ctx, t, client, app.baseURL, productA, 3)
	waitForAvailability(ctx, t, client, app.baseURL, productB, 1)

	orderInsufficientID := "f1e2d3c4-b5a6-4988-99aa-bbccddeeff22"
	orderInsufficient := newOrderCreated(orderInsufficientID, "user-2", []events.OrderLineItem{
		{ProductID: productA, Quantity: 2},
		{ProductID: productB, Quantity: 2},
	}, 2)
	publishOrderCreated(ctx, t, orderConn, orderInsufficient)
	depleted := waitForStockDepleted(ctx, t, orderConn)
	require.Equal(t, orderInsufficient.Payload.OrderID, depleted.Payload.OrderID)
	require.Len(t, depleted.Payload.Depleted, 1)
	require.Equal(t, productB, depleted.Payload.Depleted[0].ProductID)
	require.Equal(t, 2, depleted.Payload.Depleted[0].Requested)
	require.Equal(t, 1, depleted.Payload.Depleted[0].Available)
	require.Equal(t, orderInsufficient.EventEnvelope.CorrelationID, depleted.EventEnvelope.CorrelationID)

	waitForAvailability(ctx, t, client, app.baseURL, productA, 3)
	waitForAvailability(ctx, t, client, app.baseURL, productB, 1)
}

type inventoryApp struct {
	baseURL string
	stop    func()
}

func startInventoryService(ctx context.Context, t *testing.T, dbURL, rabbitURL string) *inventoryApp {
	t.Helper()

	pool, err := db.NewPool(ctx, dbURL)
	require.NoError(t, err)

	conn := dialAMQP(ctx, t, rabbitURL)

	repo := inventory.NewPostgresRepository(pool)
	logger := log.New(io.Discard, "", log.LstdFlags)

	serviceCtx, cancel := context.WithCancel(ctx)
	consumer, cleanupPub, err := events.StartOrderCreatedConsumer(serviceCtx, conn, pool, repo, logger)
	require.NoError(t, err)
	_ = consumer

	handler := httpapi.NewHandler(repo)
	router := httpapi.NewRouter(handler)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := &http.Server{
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	baseURL := fmt.Sprintf("http://%s", ln.Addr().String())

	return &inventoryApp{
		baseURL: baseURL,
		stop: func() {
			cancel()
			cleanupPub()
			_ = conn.Close()

			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			_ = server.Shutdown(shutdownCtx)
			pool.Close()

			select {
			case err := <-errCh:
				t.Logf("server error: %v", err)
			default:
			}
		},
	}
}

func startPostgres(ctx context.Context, t *testing.T) (testcontainers.Container, string) {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		Env:          map[string]string{"POSTGRES_PASSWORD": "postgres", "POSTGRES_USER": "postgres", "POSTGRES_DB": "inventory"},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/inventory?sslmode=disable", host, mappedPort.Port())
	return container, dsn
}

func startRabbitMQ(ctx context.Context, t *testing.T) (testcontainers.Container, string) {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3-management",
		ExposedPorts: []string{"5672/tcp", "15672/tcp"},
		WaitingFor:   wait.ForListeningPort("5672/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := container.MappedPort(ctx, "5672/tcp")
	require.NoError(t, err)

	return container, fmt.Sprintf("amqp://guest:guest@%s:%s/", host, mappedPort.Port())
}

func terminateContainer(t *testing.T, c testcontainers.Container) {
	t.Helper()
	terminateCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, c.Terminate(terminateCtx))
}

func seedStock(ctx context.Context, t *testing.T, client *http.Client, baseURL, productID string, available int) {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"productId": productID,
		"available": available,
	})
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/inventory/adjust", baseURL), bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func publishOrderCreated(ctx context.Context, t *testing.T, conn *amqp.Connection, order events.EnvelopedOrderCreated) {
	t.Helper()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	require.NoError(t, ch.ExchangeDeclare(events.EventsExchange, "topic", true, false, false, false, nil))

	body, err := json.Marshal(order)
	require.NoError(t, err)

	pubCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(pubCtx, events.EventsExchange, events.OrderCreatedRoutingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
	require.NoError(t, err)
}

func waitForStockReserved(ctx context.Context, t *testing.T, conn *amqp.Connection) events.StockReservedEvent {
	t.Helper()

	var ev events.StockReservedEvent
	waitForMessage(ctx, t, conn, events.StockReservedRoutingKey, &ev)
	return ev
}

func waitForStockDepleted(ctx context.Context, t *testing.T, conn *amqp.Connection) events.StockDepletedEvent {
	t.Helper()

	var ev events.StockDepletedEvent
	waitForMessage(ctx, t, conn, events.StockDepletedRoutingKey, &ev)
	return ev
}

func waitForMessage[T any](ctx context.Context, t *testing.T, conn *amqp.Connection, routingKey string, dest *T) {
	t.Helper()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	require.NoError(t, ch.ExchangeDeclare(events.EventsExchange, "topic", true, false, false, false, nil))

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)
	require.NoError(t, ch.QueueBind(q.Name, routingKey, events.EventsExchange, false, nil))

	pollCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	backoff := 50 * time.Millisecond
	for {
		select {
		case <-pollCtx.Done():
			t.Fatalf("timed out waiting for message on %s: %v", routingKey, pollCtx.Err())
		default:
		}

		msg, ok, getErr := ch.Get(q.Name, true)
		require.NoError(t, getErr)
		if ok {
			require.NoError(t, json.Unmarshal(msg.Body, dest))
			return
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

func waitForAvailability(ctx context.Context, t *testing.T, client *http.Client, baseURL, productID string, expected int) inventory.StockItem {
	t.Helper()

	pollCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	backoff := 50 * time.Millisecond
	for {
		select {
		case <-pollCtx.Done():
			t.Fatalf("timed out waiting for availability for %s: %v", productID, pollCtx.Err())
		default:
		}

		req, err := http.NewRequestWithContext(pollCtx, http.MethodGet, fmt.Sprintf("%s/api/inventory/%s", baseURL, productID), nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)

		var item inventory.StockItem
		func() {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&item))
			}
		}()

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

func dialAMQP(ctx context.Context, t *testing.T, rabbitURL string) *amqp.Connection {
	t.Helper()
	dialCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	conn, err := amqp.DialConfig(rabbitURL, amqp.Config{
		Dial: func(network, addr string) (net.Conn, error) {
			return (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Second,
			}).DialContext(dialCtx, network, addr)
		},
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
	})
	require.NoError(t, err)
	return conn
}

func newOrderCreated(orderID, userID string, items []events.OrderLineItem, seq int64) events.EnvelopedOrderCreated {
	now := time.Now().UTC()
	return events.EnvelopedOrderCreated{
		EventEnvelope: events.EventEnvelope{
			EventName:     events.EventTypeOrderCreated,
			EventVersion:  1,
			EventID:       uuid.NewString(),
			CorrelationID: uuid.NewString(),
			Producer:      "order-service",
			PartitionKey:  orderID,
			Sequence:      seq,
			OccurredAt:    now,
			Schema:        "contracts/events/order/OrderCreated.v1.payload.schema.json",
		},
		Payload: events.OrderCreatedPayload{
			OrderID:     orderID,
			UserID:      userID,
			Items:       items,
			TotalAmount: 0,
			Timestamp:   now,
		},
	}
}
