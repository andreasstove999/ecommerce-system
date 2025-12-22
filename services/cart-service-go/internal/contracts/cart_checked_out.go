package contracts

import (
	"time"

	"github.com/andreasstove999/ecommerce-system/cart-service-go/internal/cart"
	"github.com/google/uuid"
)

const (
	CartCheckedOutEventName           = "CartCheckedOut"
	CartCheckedOutEventVersion        = 1
	CartCheckedOutEnvelopedSchemaPath = "contracts/events/cart/CartCheckedOut.v1.enveloped.schema.json"
	CartServiceProducer               = "cart-service"
)

type EventEnvelope struct {
	EventName     string                `json:"eventName"`
	EventVersion  int                   `json:"eventVersion"`
	EventID       string                `json:"eventId"`
	CorrelationID string                `json:"correlationId,omitempty"`
	CausationID   string                `json:"causationId,omitempty"`
	Producer      string                `json:"producer"`
	PartitionKey  string                `json:"partitionKey"`
	Sequence      int64                 `json:"sequence"`
	OccurredAt    time.Time             `json:"occurredAt"`
	Schema        string                `json:"schema"`
	Payload       CartCheckedOutPayload `json:"payload"`
}

type CartCheckedOutPayload struct {
	CartID      string               `json:"cartId"`
	UserID      string               `json:"userId"`
	Items       []CartCheckedOutItem `json:"items"`
	TotalAmount float64              `json:"totalAmount"`
	Timestamp   time.Time            `json:"timestamp"`
}

type CartCheckedOutItem struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type EnvelopeOptions struct {
	PartitionKey  string
	Sequence      int64
	Producer      string
	SchemaPath    string
	CorrelationID string
	CausationID   string
	EventID       string
	OccurredAt    time.Time
}

func BuildCartCheckedOutEvent(c *cart.Cart, opts EnvelopeOptions) EventEnvelope {
	eventID := opts.EventID
	if eventID == "" {
		eventID = uuid.NewString()
	}

	occurredAt := opts.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	schemaPath := opts.SchemaPath
	if schemaPath == "" {
		schemaPath = CartCheckedOutEnvelopedSchemaPath
	}

	producer := opts.Producer
	if producer == "" {
		producer = CartServiceProducer
	}

	payload := CartCheckedOutPayload{
		CartID:      c.ID,
		UserID:      c.UserID,
		TotalAmount: c.Total,
		Timestamp:   occurredAt,
	}

	for _, it := range c.Items {
		payload.Items = append(payload.Items, CartCheckedOutItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			Price:     it.Price,
		})
	}

	return EventEnvelope{
		EventName:     CartCheckedOutEventName,
		EventVersion:  CartCheckedOutEventVersion,
		EventID:       eventID,
		CorrelationID: opts.CorrelationID,
		CausationID:   opts.CausationID,
		Producer:      producer,
		PartitionKey:  opts.PartitionKey,
		Sequence:      opts.Sequence,
		OccurredAt:    occurredAt,
		Schema:        schemaPath,
		Payload:       payload,
	}
}
