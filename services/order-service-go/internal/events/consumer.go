package events

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/dedup"
	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/sequence"
)

// MustDialRabbit connects to RabbitMQ or panics on failure.
func MustDialRabbit() *amqp.Connection {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@rabbitmq:5672/"
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("connect to RabbitMQ: %v", err)
	}
	return conn
}

// HandlerFunc processes a single message body.
// Return nil to ACK, return error to NACK (message will be sent to DLQ).
type HandlerFunc func(ctx context.Context, body []byte) error

// Consumer manages multiple queue subscriptions with registered handlers.
type subscription struct {
	queue      string
	routingKey string
	handler    HandlerFunc
}

type Consumer struct {
	conn   *amqp.Connection
	logger *log.Logger
	queues map[string]subscription
	dlqCh  *amqp.Channel // channel for publishing to dead letter queue
	mu     sync.RWMutex
}

// NewConsumer creates a new Consumer that will use the provided RabbitMQ connection.
func NewConsumer(conn *amqp.Connection, logger *log.Logger) *Consumer {
	// Create a dedicated channel for DLQ publishing
	dlqCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open DLQ channel: %v", err)
	}

	// Declare the dead letter queue
	_, err = dlqCh.QueueDeclare(
		"order-service.dlq", // queue name
		true,                // durable
		false,               // autoDelete
		false,               // exclusive
		false,               // noWait
		nil,                 // args
	)
	if err != nil {
		log.Fatalf("failed to declare DLQ: %v", err)
	}

	return &Consumer{
		conn:   conn,
		logger: logger,
		queues: make(map[string]subscription),
		dlqCh:  dlqCh,
	}
}

// Register associates a routing key with a handler function using the service-owned queue name.
// Must be called before Start.
func (c *Consumer) Register(routingKey string, handler HandlerFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	queue := orderQueueName(routingKey)
	c.queues[queue] = subscription{queue: queue, routingKey: routingKey, handler: handler}
}

// Start declares all registered queues and starts a consumer goroutine for each.
// It returns after all consumers are started. Cancel the context to stop all consumers.
func (c *Consumer) Start(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for queue, sub := range c.queues {
		if err := c.startQueueConsumer(ctx, sub); err != nil {
			return fmt.Errorf("start consumer for %s: %w", queue, err)
		}
		c.logger.Printf("started consumer for queue: %s", queue)
	}

	return nil
}

// startQueueConsumer creates a channel, declares the queue, and starts consuming.
func (c *Consumer) startQueueConsumer(ctx context.Context, sub subscription) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}

	if err := declareEventsExchange(ch); err != nil {
		_ = ch.Close()
		return fmt.Errorf("declare exchange: %w", err)
	}

	queue := sub.queue
	_, err = ch.QueueDeclare(
		queue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		_ = ch.Close()
		return fmt.Errorf("declare queue %s: %w", queue, err)
	}

	if err := ch.QueueBind(queue, sub.routingKey, EventsExchange, false, nil); err != nil {
		_ = ch.Close()
		return fmt.Errorf("bind queue %s: %w", queue, err)
	}

	msgs, err := ch.Consume(
		queue,
		"order-service", // consumer tag
		false,           // autoAck
		false,           // exclusive
		false,           // noLocal
		false,           // noWait
		nil,             // args
	)
	if err != nil {
		_ = ch.Close()
		return fmt.Errorf("consume %s: %w", queue, err)
	}

	go c.consumeLoop(ctx, ch, queue, msgs, sub.handler)

	return nil
}

// consumeLoop processes messages until context is cancelled or channel closes.
func (c *Consumer) consumeLoop(
	ctx context.Context,
	ch *amqp.Channel,
	queue string,
	msgs <-chan amqp.Delivery,
	handler HandlerFunc,
) {
	defer func() {
		_ = ch.Close()
		c.logger.Printf("stopped consumer for queue: %s", queue)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				c.logger.Printf("channel closed for queue: %s", queue)
				return
			}

			if err := handler(ctx, msg.Body); err != nil {
				c.logger.Printf("handler error for %s: %v", queue, err)

				// Publish to dead letter queue with error context
				if publishErr := c.publishToDLQ(ctx, queue, msg.Body, err); publishErr != nil {
					c.logger.Printf("failed to publish to DLQ for %s: %v", queue, publishErr)
				} else {
					c.logger.Printf("message from %s moved to DLQ", queue)
				}

				_ = msg.Nack(false, false) // reject without requeue
				continue
			}
			_ = msg.Ack(false)
		}
	}
}

// publishToDLQ publishes a failed message to the dead letter queue with context.
func (c *Consumer) publishToDLQ(ctx context.Context, originalQueue string, body []byte, handlerErr error) error {
	headers := amqp.Table{
		"x-original-queue": originalQueue,
		"x-error":          handlerErr.Error(),
		"x-failed-at":      time.Now().UTC().Format(time.RFC3339),
	}

	pubCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return c.dlqCh.PublishWithContext(
		pubCtx,
		"",                  // default exchange
		"order-service.dlq", // routing key (queue name)
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Headers:      headers,
		},
	)
}

// StartCartCheckedOutConsumer starts a consumer that listens for CartCheckedOut
// events and persists orders using the provided repository.
// It returns the consumer, a cleanup function for the publisher, and any error encountered.
func StartCartCheckedOutConsumer(
	ctx context.Context,
	conn *amqp.Connection,
	db *sql.DB,
	repo order.Repository,
	dedupRepo dedup.Repository,
	seqRepo sequence.Repository,
	logger *log.Logger,
	consumeEnveloped bool,
	publishEnveloped bool,
) (*Consumer, func(), error) {
	pub, err := NewPublisher(conn, seqRepo, publishEnveloped)
	if err != nil {
		return nil, nil, fmt.Errorf("create publisher: %w", err)
	}

	consumer := NewConsumer(conn, logger)
	consumer.Register(RoutingCartCheckedOut, CartCheckedOutHandler(db, repo, dedupRepo, pub, logger, consumeEnveloped))

	if err := consumer.Start(ctx); err != nil {
		_ = pub.Close()
		return nil, nil, fmt.Errorf("start consumer: %w", err)
	}

	cleanup := func() {
		_ = pub.Close()
	}

	return consumer, cleanup, nil
}
