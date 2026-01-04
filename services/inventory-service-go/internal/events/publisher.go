package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/inventory"
	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/sequence"
	"github.com/google/uuid"
)

type Publisher struct {
	ch                 *amqp.Channel
	seqRepo            *sequence.Repository
	publishEnveloped   bool
	producerIdentifier string
}

type PublisherOptions struct {
	PublishEnveloped bool
	Producer         string
}

func NewPublisher(conn *amqp.Connection, seqRepo *sequence.Repository, opts PublisherOptions) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}

	if err := declareEventsExchange(ch); err != nil {
		return nil, fmt.Errorf("declare events exchange: %w", err)
	}

	producer := opts.Producer
	if producer == "" {
		producer = "inventory-service"
	}

	return &Publisher{
		ch:                 ch,
		seqRepo:            seqRepo,
		publishEnveloped:   opts.PublishEnveloped,
		producerIdentifier: producer,
	}, nil
}

func (p *Publisher) Close() error {
	return p.ch.Close()
}

type EventMeta struct {
	CorrelationID string
	CausationID   string
	PartitionKey  string
}

func (p *Publisher) PublishStockReserved(ctx context.Context, meta EventMeta, orderID, userID string, reserved []inventory.Line) error {
	timestamp := time.Now().UTC()

	if !p.publishEnveloped {
		ev := LegacyStockReserved{
			EventType: EventTypeStockReserved,
			OrderID:   orderID,
			UserID:    userID,
			Timestamp: timestamp,
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
		return p.publishJSON(ctx, StockReservedRoutingKey, body)
	}

	payload := StockReservedPayload{
		OrderID:   orderID,
		UserID:    userID,
		Timestamp: timestamp,
	}
	for _, it := range reserved {
		payload.Items = append(payload.Items, ReservedItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
		})
	}

	seq, err := p.seqRepo.NextSequence(ctx, meta.PartitionKey)
	if err != nil {
		return fmt.Errorf("reserve sequence: %w", err)
	}

	env := newStockReservedEvent(meta, seq, p.producerIdentifier, payload, timestamp)
	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal StockReserved envelope: %w", err)
	}

	return p.publishJSON(ctx, StockReservedRoutingKey, body)
}

func (p *Publisher) PublishStockDepleted(ctx context.Context, meta EventMeta, orderID, userID string, depleted []inventory.DepletedLine, reserved []inventory.Line) error {
	timestamp := time.Now().UTC()

	if !p.publishEnveloped {
		ev := LegacyStockDepleted{
			EventType: EventTypeStockDepleted,
			OrderID:   orderID,
			UserID:    userID,
			Timestamp: timestamp,
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
		return p.publishJSON(ctx, StockDepletedRoutingKey, body)
	}

	payload := StockDepletedPayload{
		OrderID:   orderID,
		UserID:    userID,
		Timestamp: timestamp,
	}
	for _, d := range depleted {
		payload.Depleted = append(payload.Depleted, DepletedLine{
			ProductID: d.ProductID,
			Requested: d.Requested,
			Available: d.Available,
		})
	}
	for _, r := range reserved {
		payload.Reserved = append(payload.Reserved, ReservedItem{
			ProductID: r.ProductID,
			Quantity:  r.Quantity,
		})
	}

	seq, err := p.seqRepo.NextSequence(ctx, meta.PartitionKey)
	if err != nil {
		return fmt.Errorf("reserve sequence: %w", err)
	}

	env := newStockDepletedEvent(meta, seq, p.producerIdentifier, payload, timestamp)
	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal StockDepleted envelope: %w", err)
	}

	return p.publishJSON(ctx, StockDepletedRoutingKey, body)
}

func (p *Publisher) publishJSON(ctx context.Context, routingKey string, body []byte) error {
	pubCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return p.ch.PublishWithContext(
		pubCtx,
		EventsExchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

func newStockReservedEvent(meta EventMeta, seq int64, producer string, payload StockReservedPayload, occurredAt time.Time) StockReservedEvent {
	return StockReservedEvent{
		EventEnvelope: EventEnvelope{
			EventName:     EventTypeStockReserved,
			EventVersion:  1,
			EventID:       uuid.NewString(),
			CorrelationID: meta.CorrelationID,
			CausationID:   meta.CausationID,
			Producer:      producer,
			PartitionKey:  meta.PartitionKey,
			Sequence:      seq,
			OccurredAt:    occurredAt,
			Schema:        stockReservedSchema,
		},
		Payload: payload,
	}
}

func newStockDepletedEvent(meta EventMeta, seq int64, producer string, payload StockDepletedPayload, occurredAt time.Time) StockDepletedEvent {
	return StockDepletedEvent{
		EventEnvelope: EventEnvelope{
			EventName:     EventTypeStockDepleted,
			EventVersion:  1,
			EventID:       uuid.NewString(),
			CorrelationID: meta.CorrelationID,
			CausationID:   meta.CausationID,
			Producer:      producer,
			PartitionKey:  meta.PartitionKey,
			Sequence:      seq,
			OccurredAt:    occurredAt,
			Schema:        stockDepletedSchema,
		},
		Payload: payload,
	}
}
