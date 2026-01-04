package dto

import "time"

type Address struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
}

type Shipment struct {
	ShippingID     string    `json:"shippingId"`
	OrderID        string    `json:"orderId"`
	UserID         string    `json:"userId"`
	Address        Address   `json:"address"`
	ShippingMethod string    `json:"shippingMethod"`
	Carrier        string    `json:"carrier"`
	CreatedAt      time.Time `json:"createdAt"`
}
