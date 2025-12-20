package order

import "time"

type Item struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type Order struct {
	ID          string    `json:"orderId"`
	CartID      string    `json:"cartId"`
	UserID      string    `json:"userId"`
	Items       []Item    `json:"items"`
	TotalAmount float64   `json:"totalAmount"`
	CreatedAt   time.Time `json:"createdAt"`
}
