package events

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
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
type Consumer struct {
	conn   *amqp.Connection
	logger *log.Logger
	queues map[string]HandlerFunc
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
		queues: make(map[string]HandlerFunc),
		dlqCh:  dlqCh,
	}
}

// Register associates a queue name with a handler function.
// Must be called before Start.
func (c *Consumer) Register(queue string, handler HandlerFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.queues[queue] = handler
}

// Start declares all registered queues and starts a consumer goroutine for each.
// It returns after all consumers are started. Cancel the context to stop all consumers.
func (c *Consumer) Start(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for queue, handler := range c.queues {
		if err := c.startQueueConsumer(ctx, queue, handler); err != nil {
			return fmt.Errorf("start consumer for %s: %w", queue, err)
		}
		c.logger.Printf("started consumer for queue: %s", queue)
	}

	return nil
}

// startQueueConsumer creates a channel, declares the queue, and starts consuming.
func (c *Consumer) startQueueConsumer(ctx context.Context, queue string, handler HandlerFunc) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}

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

	go c.consumeLoop(ctx, ch, queue, msgs, handler)

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
