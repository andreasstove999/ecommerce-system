package events

import "time"

type CartItem struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type CartCheckedOut struct {
	EventType   string     `json:"eventType"`
	CartID      string     `json:"cartId"`
	UserID      string     `json:"userId"`
	Items       []CartItem `json:"items"`
	TotalAmount float64    `json:"totalAmount"`
	Timestamp   time.Time  `json:"timestamp"`
}
