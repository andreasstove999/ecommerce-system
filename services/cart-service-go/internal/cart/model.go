package cart

import "time"

type Item struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type Cart struct {
	ID        string    `json:"cartId"`
	UserID    string    `json:"userId"`
	Items     []Item    `json:"items"`
	Total     float64   `json:"totalAmount"`
	UpdatedAt time.Time `json:"updatedAt"`
}
