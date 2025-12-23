package events

import (
	"fmt"
	"testing"
	"time"
)

func TestStockReservedEnvelopeSchema(t *testing.T) {
	now := time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)
	meta := EventMeta{
		CorrelationID: "c0a8e2b6-3c6a-4d7e-9c8f-1f2e3d4c5b6a",
		CausationID:   "0f1e2d3c-4b5a-6978-8899-aabbccddeeff",
		PartitionKey:  "f1e2d3c4-b5a6-4988-99aa-bbccddeeff11",
	}
	payload := StockReservedPayload{
		OrderID:   meta.PartitionKey,
		UserID:    "1a2b3c4d-5e6f-7081-920a-bc0d1e2f3a4b",
		Items:     []ReservedItem{{ProductID: "123e4567-e89b-12d3-a456-426614174000", Quantity: 2}},
		Timestamp: now,
	}

	ev := newStockReservedEvent(meta, 5, "inventory-service", payload, now)
	if ev.EventName != EventTypeStockReserved || ev.EventVersion != 1 {
		t.Fatalf("unexpected name/version: %+v", ev.EventEnvelope)
	}
	if ev.PartitionKey != meta.PartitionKey {
		t.Fatalf("partition key mismatch: %s", ev.PartitionKey)
	}

	assertReservedValid(t, ev)

	// mutate to ensure validation fails
	ev.EventName = "WrongName"
	if err := validateStockReserved(ev); err == nil {
		t.Fatalf("expected validation error for wrong eventName")
	}
}

func TestStockDepletedEnvelopeSchema(t *testing.T) {
	now := time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC)
	meta := EventMeta{
		CorrelationID: "29c8f25e-2ee7-4b81-9a0d-8d3f6a0a1b2c",
		CausationID:   "59f03b7d-2d3f-4c7d-8e9f-0a1b2c3d4e5f",
		PartitionKey:  "c5a102e6-1c7b-4e5d-8a9f-0b1c2d3e4f5a",
	}
	payload := StockDepletedPayload{
		OrderID:   meta.PartitionKey,
		UserID:    "12a3b4c5-d6e7-8901-2345-6789abcdef01",
		Depleted:  []DepletedLine{{ProductID: "99887766-5544-3322-1100-aabbccddeeff", Requested: 3, Available: 1}},
		Reserved:  []ReservedItem{{ProductID: "123e4567-e89b-12d3-a456-426614174000", Quantity: 1}},
		Timestamp: now,
	}

	ev := newStockDepletedEvent(meta, 9, "inventory-service", payload, now)
	if ev.EventName != EventTypeStockDepleted || ev.EventVersion != 1 {
		t.Fatalf("unexpected name/version: %+v", ev.EventEnvelope)
	}
	if ev.CorrelationID != meta.CorrelationID || ev.CausationID != meta.CausationID {
		t.Fatalf("correlation/causation mismatch")
	}

	if err := validateStockDepleted(ev); err != nil {
		t.Fatalf("validation failed: %v", err)
	}
}

func assertReservedValid(t *testing.T, ev StockReservedEvent) {
	t.Helper()
	if err := validateStockReserved(ev); err != nil {
		t.Fatalf("schema validation failed: %v", err)
	}
}

func validateStockReserved(ev StockReservedEvent) error {
	if ev.EventName != EventTypeStockReserved {
		return errf("eventName")
	}
	if ev.EventVersion != 1 {
		return errf("eventVersion")
	}
	if ev.Schema != stockReservedSchema {
		return errf("schema")
	}
	if ev.Producer == "" || ev.PartitionKey == "" || ev.EventID == "" {
		return errf("envelope required")
	}
	if ev.Payload.OrderID == "" || ev.Payload.UserID == "" || len(ev.Payload.Items) == 0 || ev.Payload.Timestamp.IsZero() {
		return errf("payload required")
	}
	for _, it := range ev.Payload.Items {
		if it.ProductID == "" || it.Quantity <= 0 {
			return errf("invalid item")
		}
	}
	return nil
}

func validateStockDepleted(ev StockDepletedEvent) error {
	if ev.EventName != EventTypeStockDepleted || ev.EventVersion != 1 {
		return errf("envelope fields")
	}
	if ev.Schema != stockDepletedSchema || ev.EventID == "" || ev.PartitionKey == "" {
		return errf("envelope required")
	}
	if ev.Payload.OrderID == "" || ev.Payload.UserID == "" || len(ev.Payload.Depleted) == 0 || ev.Payload.Timestamp.IsZero() {
		return errf("payload required")
	}
	for _, d := range ev.Payload.Depleted {
		if d.ProductID == "" || d.Requested <= 0 || d.Available < 0 {
			return errf("invalid depleted")
		}
	}
	for _, r := range ev.Payload.Reserved {
		if r.ProductID == "" || r.Quantity <= 0 {
			return errf("invalid reserved")
		}
	}
	return nil
}

func errf(field string) error {
	return fmt.Errorf("validation failed for %s", field)
}
