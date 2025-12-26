package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/sequence"
)

// TODO compare with cart-service-go/internal/events/rabbit.go there seems to be a lot of code duplication and some mismatches

const (
	OrderCreatedQueue   = "order.created"
	OrderCompletedQueue = "order.completed"
)

type Publisher struct {
	ch               *amqp.Channel
	seqRepo          sequence.Repository
	publishEnveloped bool
}

func NewPublisher(conn *amqp.Connection, seqRepo sequence.Repository, publishEnveloped bool) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}

	// Declare queues so publish never fails due to missing infra
	_, err = ch.QueueDeclare(OrderCreatedQueue, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("declare %s: %w", OrderCreatedQueue, err)
	}

	return &Publisher{
		ch:               ch,
		seqRepo:          seqRepo,
		publishEnveloped: publishEnveloped,
	}, nil
}

func (p *Publisher) Close() error {
	return p.ch.Close()
}

func (p *Publisher) PublishOrderCreated(ctx context.Context, o *order.Order, meta EnvelopeMetadata) error {
	if !p.publishEnveloped {
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
			return fmt.Errorf("marshal OrderCreated legacy: %w", err)
		}
		return p.publishJSON(ctx, OrderCreatedQueue, body)
	}

	seq, err := p.seqRepo.NextSequence(ctx, o.ID)
	if err != nil {
		return fmt.Errorf("next sequence: %w", err)
	}

	env := BuildOrderCreatedEnvelope(o, seq, meta)
	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal OrderCreated enveloped: %w", err)
	}

	return p.publishJSON(ctx, OrderCreatedQueue, body)
}

func (p *Publisher) PublishOrderCompleted(ctx context.Context, orderID, userID string, meta EnvelopeMetadata) error {
	if !p.publishEnveloped {
		ev := OrderCompleted{
			EventType: "OrderCompleted",
			OrderID:   orderID,
			UserID:    userID,
			Timestamp: time.Now().UTC(),
		}

		body, err := json.Marshal(ev)
		if err != nil {
			return fmt.Errorf("marshal OrderCompleted legacy: %w", err)
		}

		return p.publishJSON(ctx, OrderCompletedQueue, body)
	}

	seq, err := p.seqRepo.NextSequence(ctx, orderID)
	if err != nil {
		return fmt.Errorf("next sequence: %w", err)
	}

	env := BuildOrderCompletedEnvelope(orderID, userID, seq, meta)
	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal OrderCompleted enveloped: %w", err)
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
