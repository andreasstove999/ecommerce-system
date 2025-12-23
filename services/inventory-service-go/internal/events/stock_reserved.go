package events

import (
	"time"
)

const EventTypeStockReserved = "StockReserved"

const stockReservedSchema = "contracts/events/inventory/StockReserved.v1.enveloped.schema.json"

type StockReservedPayload struct {
	OrderID   string         `json:"orderId"`
	UserID    string         `json:"userId"`
	Items     []ReservedItem `json:"items"`
	Timestamp time.Time      `json:"timestamp"`
}

type ReservedItem struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type StockReservedEvent struct {
	EventEnvelope
	Payload StockReservedPayload `json:"payload"`
}

// LegacyStockReserved matches the pre-envelope payload for backward compatibility.
type LegacyStockReserved struct {
	EventType string      `json:"eventType"`
	OrderID   string      `json:"orderId"`
	UserID    string      `json:"userId"`
	Timestamp time.Time   `json:"timestamp"`
	Items     []StockLine `json:"items"`
}

type StockLine struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}
