package dto

import "time"

type OrderItem struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type Order struct {
	OrderID     string      `json:"orderId"`
	CartID      string      `json:"cartId"`
	UserID      string      `json:"userId"`
	Items       []OrderItem `json:"items"`
	TotalAmount float64     `json:"totalAmount"`
	CreatedAt   time.Time   `json:"createdAt"`
}
