package events

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	// QueueOrderCreated is published by order-service-go Publisher.PublishOrderCreated.
	QueueOrderCreated = "order.created"

	EventTypeOrderCreated = "OrderCreated"
)

// OrderCreatedPayload matches the v1 payload schema.
type OrderCreatedPayload struct {
	OrderID     string          `json:"orderId"`
	CartID      string          `json:"cartId,omitempty"`
	UserID      string          `json:"userId"`
	Items       []OrderLineItem `json:"items"`
	TotalAmount float64         `json:"totalAmount,omitempty"`
	Timestamp   time.Time       `json:"timestamp"`
}

// EnvelopedOrderCreated represents the incoming v1 enveloped message.
type EnvelopedOrderCreated struct {
	EventEnvelope
	Payload OrderCreatedPayload `json:"payload"`
}

type OrderLineItem struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price,omitempty"`
}

// legacyOrderCreated represents the previous non-enveloped shape.
type legacyOrderCreated struct {
	EventType   string          `json:"eventType"`
	OrderID     string          `json:"orderId"`
	CartID      string          `json:"cartId,omitempty"`
	UserID      string          `json:"userId"`
	Items       []OrderLineItem `json:"items"`
	TotalAmount float64         `json:"totalAmount,omitempty"`
	Timestamp   time.Time       `json:"timestamp"`
}

type OrderCreatedMessage struct {
	Envelope *EventEnvelope
	Payload  OrderCreatedPayload
	Legacy   bool
}

func parseOrderCreated(body []byte, consumeEnveloped bool) (OrderCreatedMessage, error) {
	if consumeEnveloped {
		env, err := parseEnvelope(body)
		if err == nil && env.EventName != "" {
			if err := env.Validate(EventTypeOrderCreated, 1); err != nil {
				return OrderCreatedMessage{}, fmt.Errorf("envelope validate: %w", err)
			}
			var payload OrderCreatedPayload
			if err := json.Unmarshal(env.Payload, &payload); err != nil {
				return OrderCreatedMessage{}, fmt.Errorf("unmarshal order payload: %w", err)
			}
			return OrderCreatedMessage{Envelope: &env, Payload: payload}, nil
		}
	}

	var legacy legacyOrderCreated
	if err := json.Unmarshal(body, &legacy); err != nil {
		return OrderCreatedMessage{}, fmt.Errorf("unmarshal legacy order: %w", err)
	}
	if legacy.OrderID == "" {
		return OrderCreatedMessage{}, fmt.Errorf("missing orderId")
	}

	return OrderCreatedMessage{
		Payload: OrderCreatedPayload{
			OrderID:     legacy.OrderID,
			CartID:      legacy.CartID,
			UserID:      legacy.UserID,
			Items:       legacy.Items,
			TotalAmount: legacy.TotalAmount,
			Timestamp:   legacy.Timestamp,
		},
		Legacy: true,
	}, nil
}
