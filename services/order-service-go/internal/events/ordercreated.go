package events

import (
	"time"

	"github.com/google/uuid"

	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
)

const (
	orderCreatedEventName    = "OrderCreated"
	orderCreatedEventVersion = 1
	orderCreatedSchema       = "contracts/events/order/OrderCreated.v1.payload.schema.json"
)

// OrderCreatedPayload represents the v1 payload schema.
type OrderCreatedPayload struct {
	OrderID     string     `json:"orderId"`
	CartID      string     `json:"cartId"`
	UserID      string     `json:"userId"`
	Items       []CartItem `json:"items"`
	TotalAmount float64    `json:"totalAmount"`
	Timestamp   time.Time  `json:"timestamp"`
}

// OrderCreatedEnvelope is the enveloped event structure.
type OrderCreatedEnvelope = EventEnvelope[OrderCreatedPayload]

// BuildOrderCreatedEnvelope builds an enveloped OrderCreated event.
func BuildOrderCreatedEnvelope(o *order.Order, seq int64, meta EnvelopeMetadata) OrderCreatedEnvelope {
	if meta.CorrelationID == "" {
		meta.CorrelationID = uuid.NewString()
	}

	items := make([]CartItem, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, CartItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			Price:     it.Price,
		})
	}

	return OrderCreatedEnvelope{
		EventName:     orderCreatedEventName,
		EventVersion:  orderCreatedEventVersion,
		EventID:       uuid.NewString(),
		CorrelationID: meta.CorrelationID,
		CausationID:   meta.CausationID,
		Producer:      "order-service",
		PartitionKey:  o.ID,
		Sequence:      &seq,
		OccurredAt:    time.Now().UTC(),
		Schema:        orderCreatedSchema,
		Payload: OrderCreatedPayload{
			OrderID:     o.ID,
			CartID:      o.CartID,
			UserID:      o.UserID,
			Items:       items,
			TotalAmount: o.TotalAmount,
			Timestamp:   o.CreatedAt,
		},
	}
}
