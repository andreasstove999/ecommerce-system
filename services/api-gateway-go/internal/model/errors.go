package model

type ErrorResponse struct {
	Error         string `json:"error"`
	CorrelationID string `json:"correlationId,omitempty"`
}
