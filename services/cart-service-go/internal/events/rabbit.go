package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/andreasstove999/ecommerce-system/cart-service-go/internal/cart"
	amqp "github.com/rabbitmq/amqp091-go"
)

const cartCheckedOutQueue = "cart.checkedout"

type RabbitCartEventsPublisher struct {
	ch *amqp.Channel
}

func NewRabbitCartEventsPublisher(conn *amqp.Connection) (*RabbitCartEventsPublisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}

	_, err = ch.QueueDeclare(
		cartCheckedOutQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %v", err)
	}
	return &RabbitCartEventsPublisher{ch: ch}, nil
}

func MustDialRabbit() *amqp.Connection {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@rabbitmq:5672/"
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	return conn
}

func (p *RabbitCartEventsPublisher) PublishCartCheckedOut(ctx context.Context, c *cart.Cart) error {
	ev := CartCheckedOut{
		EventType:   "CartCheckedOut",
		CartID:      c.ID,
		UserID:      c.UserID,
		TotalAmount: c.Total,
		Timestamp:   time.Now().UTC(),
	}

	for _, it := range c.Items {
		ev.Items = append(ev.Items, CartItemEvent{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			Price:     it.Price,
		})
	}

	body, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	pub := amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	}

	// Empty exchange = default; routingKey = queue name
	if err := p.ch.PublishWithContext(
		ctx,
		"",                  // exchange
		cartCheckedOutQueue, // routing key
		false,
		false,
		pub,
	); err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	return nil
}
