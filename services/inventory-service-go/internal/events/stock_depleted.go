package events

import "time"

const EventTypeStockDepleted = "StockDepleted"

const stockDepletedSchema = "contracts/events/inventory/StockDepleted.v1.enveloped.schema.json"

type StockDepletedPayload struct {
	OrderID   string         `json:"orderId"`
	UserID    string         `json:"userId"`
	Depleted  []DepletedLine `json:"depleted"`
	Reserved  []ReservedItem `json:"reserved,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

type StockDepletedEvent struct {
	EventEnvelope
	Payload StockDepletedPayload `json:"payload"`
}

// LegacyStockDepleted matches the pre-envelope payload for backward compatibility.
type LegacyStockDepleted struct {
	EventType string         `json:"eventType"`
	OrderID   string         `json:"orderId"`
	UserID    string         `json:"userId"`
	Timestamp time.Time      `json:"timestamp"`
	Depleted  []DepletedLine `json:"depleted"`
	Reserved  []StockLine    `json:"reserved,omitempty"`
}

type DepletedLine struct {
	ProductID string `json:"productId"`
	Requested int    `json:"requested"`
	Available int    `json:"available"`
}
