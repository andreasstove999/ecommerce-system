package dto

import "time"

type Payment struct {
	PaymentID     string    `json:"paymentId"`
	OrderID       string    `json:"orderId"`
	UserID        string    `json:"userId"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        int       `json:"status"`
	Provider      string    `json:"provider"`
	FailureReason *string   `json:"failureReason,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}
