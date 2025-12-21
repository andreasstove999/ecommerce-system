package events

import "time"

const EventTypeStockDepleted = "StockDepleted"

type StockDepleted struct {
	EventType  string         `json:"eventType"`
	OrderID    string         `json:"orderId"`
	UserID     string         `json:"userId"`
	Timestamp  time.Time      `json:"timestamp"`
	Depleted   []DepletedLine `json:"depleted"`
	Reserved   []StockLine    `json:"reserved,omitempty"`
}

type DepletedLine struct {
	ProductID string `json:"productId"`
	Requested int    `json:"requested"`
	Available int    `json:"available"`
}
