package dto

type AvailabilityResponse struct {
	ProductID string `json:"productId"`
	Available int    `json:"available"`
}

type AdjustInventoryRequest struct {
	ProductID string `json:"productId"`
	Available int    `json:"available"`
}

type AdjustInventoryResponse struct {
	OK bool `json:"ok"`
}
