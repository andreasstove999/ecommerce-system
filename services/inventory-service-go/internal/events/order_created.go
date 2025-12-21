package events

import "time"

const (
	// QueueOrderCreated is published by order-service-go Publisher.PublishOrderCreated.
	QueueOrderCreated = "order.created"

	EventTypeOrderCreated = "OrderCreated"
)

// OrderCreated is published by order-service.
// Inventory consumes this and attempts to reserve stock.
type OrderCreated struct {
	EventType    string     `json:"eventType"`
	OrderID      string     `json:"orderId"`
	CartID       string     `json:"cartId,omitempty"`
	UserID       string     `json:"userId"`
	TotalAmount  float64    `json:"totalAmount,omitempty"`
	Timestamp    time.Time  `json:"timestamp"`
	Items        []CartItem `json:"items"`
}

// CartItem matches the cart/order item contract used across services.
type CartItem struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price,omitempty"`
}
