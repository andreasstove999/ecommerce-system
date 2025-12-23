package events

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	consumeEnvelopedEnv = "CONSUME_ENVELOPED_EVENTS"
	publishEnvelopedEnv = "PUBLISH_ENVELOPED_EVENTS"
)

// EventEnvelope represents the shared envelope for v1 contracts.
type EventEnvelope struct {
	EventName     string          `json:"eventName"`
	EventVersion  int             `json:"eventVersion"`
	EventID       string          `json:"eventId"`
	CorrelationID string          `json:"correlationId,omitempty"`
	CausationID   string          `json:"causationId,omitempty"`
	Producer      string          `json:"producer"`
	PartitionKey  string          `json:"partitionKey"`
	Sequence      int64           `json:"sequence,omitempty"`
	OccurredAt    time.Time       `json:"occurredAt"`
	Schema        string          `json:"schema"`
	Payload       json.RawMessage `json:"payload"`
}

func (e EventEnvelope) Validate(expectedName string, expectedVersion int) error {
	if e.EventName != expectedName {
		return fmt.Errorf("unexpected eventName %q", e.EventName)
	}
	if e.EventVersion != expectedVersion {
		return fmt.Errorf("unexpected eventVersion %d", e.EventVersion)
	}
	if e.PartitionKey == "" {
		return fmt.Errorf("missing partitionKey")
	}
	if e.EventID == "" {
		return fmt.Errorf("missing eventId")
	}
	return nil
}

type envelopeDecoder struct {
	EventName     string          `json:"eventName"`
	EventVersion  int             `json:"eventVersion"`
	EventID       string          `json:"eventId"`
	CorrelationID string          `json:"correlationId,omitempty"`
	CausationID   string          `json:"causationId,omitempty"`
	Producer      string          `json:"producer"`
	PartitionKey  string          `json:"partitionKey"`
	Sequence      int64           `json:"sequence,omitempty"`
	OccurredAt    time.Time       `json:"occurredAt"`
	Schema        string          `json:"schema"`
	Payload       json.RawMessage `json:"payload"`
}

func parseEnvelope(body []byte) (EventEnvelope, error) {
	var decoded envelopeDecoder
	if err := json.Unmarshal(body, &decoded); err != nil {
		return EventEnvelope{}, err
	}
	return EventEnvelope{
		EventName:     decoded.EventName,
		EventVersion:  decoded.EventVersion,
		EventID:       decoded.EventID,
		CorrelationID: decoded.CorrelationID,
		CausationID:   decoded.CausationID,
		Producer:      decoded.Producer,
		PartitionKey:  decoded.PartitionKey,
		Sequence:      decoded.Sequence,
		OccurredAt:    decoded.OccurredAt,
		Schema:        decoded.Schema,
		Payload:       decoded.Payload,
	}, nil
}
