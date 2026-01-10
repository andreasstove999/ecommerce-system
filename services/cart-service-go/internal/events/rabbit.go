package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/andreasstove999/ecommerce-system/cart-service-go/internal/cart"
	"github.com/andreasstove999/ecommerce-system/cart-service-go/internal/contracts"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SequenceRepository interface {
	NextSequence(ctx context.Context, partitionKey string) (int64, error)
}

type RabbitCartEventsPublisher struct {
	ch                 *amqp.Channel
	sequenceRepository SequenceRepository
	publishEnveloped   bool
}

func (p *RabbitCartEventsPublisher) Close() error {
	if p == nil || p.ch == nil {
		return nil
	}
	return p.ch.Close()
}

func NewRabbitCartEventsPublisher(conn *amqp.Connection, sequenceRepo SequenceRepository) (*RabbitCartEventsPublisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}

	if err := declareEventsExchange(ch); err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %v", err)
	}

	publishEnveloped := true
	if v := os.Getenv("PUBLISH_ENVELOPED_EVENTS"); strings.EqualFold(v, "false") {
		publishEnveloped = false
	}

	return &RabbitCartEventsPublisher{
		ch:                 ch,
		sequenceRepository: sequenceRepo,
		publishEnveloped:   publishEnveloped,
	}, nil
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

func (p *RabbitCartEventsPublisher) PublishCartCheckedOut(ctx context.Context, c *cart.Cart, metadata PublishMetadata) error {
	if p.publishEnveloped {
		return p.publishEnvelopedEvent(ctx, c, metadata)
	}
	return p.publishLegacyEvent(ctx, c)
}

func (p *RabbitCartEventsPublisher) publishLegacyEvent(ctx context.Context, c *cart.Cart) error {
	ev := LegacyCartCheckedOut{
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
		return fmt.Errorf("marshal legacy event: %w", err)
	}

	pub := amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	}

	if err := p.ch.PublishWithContext(ctx, EventsExchange, CartCheckedOutRoutingKey, false, false, pub); err != nil {
		return fmt.Errorf("publish legacy: %w", err)
	}

	return nil
}

func (p *RabbitCartEventsPublisher) publishEnvelopedEvent(ctx context.Context, c *cart.Cart, metadata PublishMetadata) error {
	if p.sequenceRepository == nil {
		return fmt.Errorf("sequence repository not configured")
	}

	partitionKey := c.ID
	if partitionKey == "" {
		partitionKey = c.UserID
	}

	sequence, err := p.sequenceRepository.NextSequence(ctx, partitionKey)
	if err != nil {
		return fmt.Errorf("next sequence: %w", err)
	}

	envelope := contracts.BuildCartCheckedOutEvent(c, contracts.EnvelopeOptions{
		PartitionKey:  partitionKey,
		Sequence:      sequence,
		Producer:      contracts.CartServiceProducer,
		SchemaPath:    contracts.CartCheckedOutEnvelopedSchemaPath,
		CorrelationID: metadata.CorrelationID,
		CausationID:   metadata.CausationID,
	})

	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal enveloped event: %w", err)
	}

	pub := amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	}

	if err := p.ch.PublishWithContext(ctx, EventsExchange, CartCheckedOutRoutingKey, false, false, pub); err != nil {
		return fmt.Errorf("publish enveloped: %w", err)
	}

	return nil
}
