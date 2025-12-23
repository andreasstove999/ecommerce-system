package events

import (
	"time"

	"github.com/google/uuid"
)

type OrderCompleted struct {
	EventType string    `json:"eventType"`
	OrderID   string    `json:"orderId"`
	UserID    string    `json:"userId"`
	Timestamp time.Time `json:"timestamp"`
}

type OrderCompletedPayload struct {
	OrderID   string    `json:"orderId"`
	UserID    string    `json:"userId"`
	Timestamp time.Time `json:"timestamp"`
}

type OrderCompletedEnvelope = EventEnvelope[OrderCompletedPayload]

func BuildOrderCompletedEnvelope(orderID, userID string, seq int64, meta EnvelopeMetadata) OrderCompletedEnvelope {
	return OrderCompletedEnvelope{
		EventName:     "OrderCompleted",
		EventVersion:  1,
		EventID:       uuid.NewString(),
		CorrelationID: meta.CorrelationID,
		CausationID:   meta.CausationID,
		Producer:      "order-service",
		PartitionKey:  orderID,
		Sequence:      &seq,
		OccurredAt:    time.Now().UTC(),
		Schema:        "contracts/events/order/OrderCompleted.v1.payload.schema.json",
		Payload: OrderCompletedPayload{
			OrderID:   orderID,
			UserID:    userID,
			Timestamp: time.Now().UTC(),
		},
	}
}
