package events

import (
	"context"
	"time"

	"github.com/andreasstove999/ecommerce-system/cart-service-go/internal/cart"
)

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

type CartEventsPublisher interface {
	PublishCartCheckedOut(ctx context.Context, c *cart.Cart) error
}
