package events

import "time"

type OrderCompleted struct {
	EventType string    `json:"eventType"`
	OrderID   string    `json:"orderId"`
	UserID    string    `json:"userId"`
	Timestamp time.Time `json:"timestamp"`
}
