package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
)

// TODO compare with cart-service-go/internal/events/rabbit.go there seems to be a lot of code duplication and some mismatches

const (
	OrderCreatedQueue   = "order.created"
	OrderCompletedQueue = "order.completed"
)

type Publisher struct {
	ch *amqp.Channel
}

func NewPublisher(conn *amqp.Connection) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}

	// Declare queues so publish never fails due to missing infra
	_, err = ch.QueueDeclare(OrderCreatedQueue, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("declare %s: %w", OrderCreatedQueue, err)
	}
	_, err = ch.QueueDeclare(OrderCompletedQueue, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("declare %s: %w", OrderCompletedQueue, err)
	}

	return &Publisher{ch: ch}, nil
}

func (p *Publisher) Close() error {
	return p.ch.Close()
}

func (p *Publisher) PublishOrderCreated(ctx context.Context, o *order.Order) error {
	ev := OrderCreated{
		EventType:   "OrderCreated",
		OrderID:     o.ID,
		CartID:      o.CartID,
		UserID:      o.UserID,
		TotalAmount: o.TotalAmount,
		Timestamp:   time.Now().UTC(),
	}

	// Reuse CartItem contract so events are consistent across services
	for _, it := range o.Items {
		ev.Items = append(ev.Items, CartItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			Price:     it.Price,
		})
	}

	body, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshal OrderCreated: %w", err)
	}

	return p.publishJSON(ctx, OrderCreatedQueue, body)
}

func (p *Publisher) PublishOrderCompleted(ctx context.Context, orderID, userID string) error {
	ev := OrderCompleted{
		EventType: "OrderCompleted",
		OrderID:   orderID,
		UserID:    userID,
		Timestamp: time.Now().UTC(),
	}

	body, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshal OrderCompleted: %w", err)
	}

	return p.publishJSON(ctx, OrderCompletedQueue, body)
}

func (p *Publisher) publishJSON(ctx context.Context, routingKey string, body []byte) error {
	pubCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return p.ch.PublishWithContext(
		pubCtx,
		"",         // default exchange
		routingKey, // queue name as routing key
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}
