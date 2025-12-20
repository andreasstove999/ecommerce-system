package events

import "time"

type PaymentSucceeded struct {
	EventType string    `json:"eventType"`
	OrderID   string    `json:"orderId"`
	UserID    string    `json:"userId"`
	Timestamp time.Time `json:"timestamp"`
}

type PaymentFailed struct {
	EventType string    `json:"eventType"`
	OrderID   string    `json:"orderId"`
	UserID    string    `json:"userId"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// TODO collect all events in one file or make a separate package for events
