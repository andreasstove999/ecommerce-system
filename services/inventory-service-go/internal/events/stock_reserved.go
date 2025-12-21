package events

import "time"

const EventTypeStockReserved = "StockReserved"

type StockReserved struct {
	EventType  string      `json:"eventType"`
	OrderID    string      `json:"orderId"`
	UserID     string      `json:"userId"`
	Timestamp  time.Time   `json:"timestamp"`
	Items      []StockLine `json:"items"`
}

type StockLine struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}
