package events

import "time"

type OrderCreated struct {
	EventType   string     `json:"eventType"`
	OrderID     string     `json:"orderId"`
	CartID      string     `json:"cartId"`
	UserID      string     `json:"userId"`
	Items       []CartItem `json:"items"`
	TotalAmount float64    `json:"totalAmount"`
	Timestamp   time.Time  `json:"timestamp"`
}
