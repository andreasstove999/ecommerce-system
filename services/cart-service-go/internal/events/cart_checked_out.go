package events

import "time"

type CartCheckedOut struct {
	EventType   string          `json:"eventType"`
	CartID      string          `json:"cartId"`
	UserID      string          `json:"userId"`
	Items       []CartItemEvent `json:"items"`
	TotalAmount float64         `json:"totalAmount"`
	Timestamp   time.Time       `json:"timestamp"`
}

type CartItemEvent struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}
