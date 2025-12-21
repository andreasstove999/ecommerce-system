package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/inventory"
)

const (
	StockReservedQueue = "stock.reserved"
	StockDepletedQueue = "stock.depleted"
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
	_, err = ch.QueueDeclare(StockReservedQueue, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("declare %s: %w", StockReservedQueue, err)
	}
	_, err = ch.QueueDeclare(StockDepletedQueue, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("declare %s: %w", StockDepletedQueue, err)
	}

	return &Publisher{ch: ch}, nil
}

func (p *Publisher) Close() error {
	return p.ch.Close()
}

func (p *Publisher) PublishStockReserved(ctx context.Context, orderID, userID string, reserved []inventory.Line) error {
	ev := StockReserved{
		EventType:  EventTypeStockReserved,
		OrderID:    orderID,
		UserID:     userID,
		Timestamp:  time.Now().UTC(),
	}

	for _, it := range reserved {
		ev.Items = append(ev.Items, StockLine{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
		})
	}

	body, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshal StockReserved: %w", err)
	}

	return p.publishJSON(ctx, StockReservedQueue, body)
}

func (p *Publisher) PublishStockDepleted(ctx context.Context, orderID, userID string, depleted []inventory.DepletedLine, reserved []inventory.Line) error {
	ev := StockDepleted{
		EventType: EventTypeStockDepleted,
		OrderID:   orderID,
		UserID:    userID,
		Timestamp: time.Now().UTC(),
	}

	for _, d := range depleted {
		ev.Depleted = append(ev.Depleted, DepletedLine{
			ProductID: d.ProductID,
			Requested: d.Requested,
			Available: d.Available,
		})
	}
	for _, r := range reserved {
		ev.Reserved = append(ev.Reserved, StockLine{
			ProductID: r.ProductID,
			Quantity:  r.Quantity,
		})
	}

	body, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshal StockDepleted: %w", err)
	}

	return p.publishJSON(ctx, StockDepletedQueue, body)
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
