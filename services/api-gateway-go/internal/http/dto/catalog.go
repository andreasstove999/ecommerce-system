package dto

import "time"

type CreateProductRequest struct {
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency,omitempty"`
}

type Product struct {
	ID          string    `json:"id"`
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Currency    string    `json:"currency"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
