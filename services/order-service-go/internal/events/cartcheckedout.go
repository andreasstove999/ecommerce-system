package events

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	cartCheckedOutEventName    = "CartCheckedOut"
	cartCheckedOutEventVersion = 1
)

// CartCheckedOutPayload represents the v1 payload schema.
type CartCheckedOutPayload struct {
	CartID      string     `json:"cartId"`
	UserID      string     `json:"userId"`
	Items       []CartItem `json:"items"`
	TotalAmount float64    `json:"totalAmount"`
	Timestamp   time.Time  `json:"timestamp"`
}

// CartCheckedOutEnvelope is the enveloped event structure.
type CartCheckedOutEnvelope = EventEnvelope[CartCheckedOutPayload]

// parseCartCheckedOut parses an incoming CartCheckedOut message.
// If allowEnveloped is true it will first try the v1 envelope format and
// fall back to the legacy bare payload if unmarshalling fails.
func parseCartCheckedOut(body []byte, allowEnveloped bool) (CartCheckedOutPayload, *CartCheckedOutEnvelope, error) {
	if allowEnveloped {
		var env CartCheckedOutEnvelope
		if err := json.Unmarshal(body, &env); err == nil && env.EventName != "" {
			if err := env.Validate(cartCheckedOutEventName, cartCheckedOutEventVersion); err != nil {
				return CartCheckedOutPayload{}, nil, fmt.Errorf("invalid envelope: %w", err)
			}
			if env.Payload.CartID == "" || env.Payload.UserID == "" {
				return CartCheckedOutPayload{}, nil, fmt.Errorf("invalid payload: missing cartId or userId")
			}
			return env.Payload, &env, nil
		}
	}

	var legacy CartCheckedOut
	if err := json.Unmarshal(body, &legacy); err != nil {
		return CartCheckedOutPayload{}, nil, fmt.Errorf("unmarshal legacy CartCheckedOut: %w", err)
	}

	payload := CartCheckedOutPayload{
		CartID:      legacy.CartID,
		UserID:      legacy.UserID,
		Items:       legacy.Items,
		TotalAmount: legacy.TotalAmount,
		Timestamp:   legacy.Timestamp,
	}
	if payload.CartID == "" || payload.UserID == "" {
		return CartCheckedOutPayload{}, nil, fmt.Errorf("invalid payload: missing cartId or userId")
	}

	return payload, nil, nil
}
