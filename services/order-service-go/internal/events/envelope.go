package events

import (
	"fmt"
	"time"
)

// EventEnvelope represents the common envelope for all events.
// It is generic to allow strongly typed payloads per event.
type EventEnvelope[T any] struct {
	EventName     string    `json:"eventName"`
	EventVersion  int       `json:"eventVersion"`
	EventID       string    `json:"eventId"`
	CorrelationID string    `json:"correlationId,omitempty"`
	CausationID   string    `json:"causationId,omitempty"`
	Producer      string    `json:"producer"`
	PartitionKey  string    `json:"partitionKey"`
	Sequence      *int64    `json:"sequence,omitempty"`
	OccurredAt    time.Time `json:"occurredAt"`
	Schema        string    `json:"schema"`
	Payload       T         `json:"payload"`
}

// EnvelopeMetadata carries correlation/causation context for emitted events.
type EnvelopeMetadata struct {
	CorrelationID string
	CausationID   string
}

// Validate ensures the envelope contains the expected event identity.
func (e EventEnvelope[T]) Validate(expectedName string, expectedVersion int) error {
	if e.EventName != expectedName {
		return fmt.Errorf("unexpected eventName: %s", e.EventName)
	}
	if e.EventVersion != expectedVersion {
		return fmt.Errorf("unexpected eventVersion: %d", e.EventVersion)
	}
	if e.PartitionKey == "" {
		return fmt.Errorf("missing partitionKey")
	}
	return nil
}
