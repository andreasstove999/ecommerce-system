package dto

import "time"

type CartItem struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type Cart struct {
	CartID      string     `json:"cartId"`
	UserID      string     `json:"userId"`
	Items       []CartItem `json:"items"`
	TotalAmount float64    `json:"totalAmount"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type AddCartItemRequest struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type CheckoutResponse struct {
	Status string `json:"status"`
}
