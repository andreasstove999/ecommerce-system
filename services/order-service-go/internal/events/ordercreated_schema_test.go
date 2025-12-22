package events

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/order"
)

func TestOrderCreatedEnvelopeSchema(t *testing.T) {
	envelopeSchema := loadSchema(t, "OrderCreated.v1.enveloped.schema.json")
	payloadSchema := loadSchema(t, "OrderCreated.v1.payload.schema.json")

	o := &order.Order{
		ID:          uuid.NewString(),
		CartID:      uuid.NewString(),
		UserID:      uuid.NewString(),
		TotalAmount: 25.50,
		CreatedAt:   time.Now().UTC(),
		Items: []order.Item{
			{ProductID: uuid.NewString(), Quantity: 1, Price: 10.00},
			{ProductID: uuid.NewString(), Quantity: 2, Price: 7.75},
		},
	}

	validate := func(env OrderCreatedEnvelope) error {
		var asMap map[string]interface{}
		body, err := json.Marshal(env)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(body, &asMap); err != nil {
			return err
		}
		for _, field := range requiredFields(envelopeSchema) {
			if _, ok := asMap[field]; !ok {
				return fmt.Errorf("missing required field %s", field)
			}
		}
		if err := assertConst(envelopeSchema, asMap, "eventName"); err != nil {
			return err
		}
		if err := assertConst(envelopeSchema, asMap, "eventVersion"); err != nil {
			return err
		}
		if err := assertConst(envelopeSchema, asMap, "schema"); err != nil {
			return err
		}

		payloadMap, ok := asMap["payload"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("missing payload object")
		}
		for _, field := range requiredFields(payloadSchema) {
			if _, ok := payloadMap[field]; !ok {
				return fmt.Errorf("missing payload field %s", field)
			}
		}
		return nil
	}

	env := BuildOrderCreatedEnvelope(o, 1, EnvelopeMetadata{
		CorrelationID: uuid.NewString(),
		CausationID:   uuid.NewString(),
	})
	require.NoError(t, validate(env))

	env.EventName = "WrongEvent"
	require.Error(t, validate(env))
}

func loadSchema(t *testing.T, filename string) map[string]interface{} {
	t.Helper()
	schemaPath := filepath.Join("..", "..", "..", "..", "contracts", "events", "order", filename)
	raw, err := os.ReadFile(schemaPath)
	require.NoError(t, err)

	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &parsed))
	return parsed
}

func requiredFields(schema map[string]interface{}) []string {
	seen := make(map[string]struct{})
	addFields := func(list []interface{}) {
		for _, f := range list {
			if s, ok := f.(string); ok {
				seen[s] = struct{}{}
			}
		}
	}

	if req, ok := schema["required"].([]interface{}); ok {
		addFields(req)
	}

	if allOf, ok := schema["allOf"].([]interface{}); ok {
		for _, part := range allOf {
			if partMap, ok := part.(map[string]interface{}); ok {
				if req, ok := partMap["required"].([]interface{}); ok {
					addFields(req)
				}
			}
		}
	}

	fields := make([]string, 0, len(seen))
	for field := range seen {
		fields = append(fields, field)
	}
	return fields
}

func constValue(schema map[string]interface{}, key string) (interface{}, bool) {
	if props, ok := schema["properties"].(map[string]interface{}); ok {
		if prop, ok := props[key].(map[string]interface{}); ok {
			if val, ok := prop["const"]; ok {
				return val, true
			}
		}
	}

	if allOf, ok := schema["allOf"].([]interface{}); ok {
		for _, part := range allOf {
			if partMap, ok := part.(map[string]interface{}); ok {
				if val, found := constValue(partMap, key); found {
					return val, true
				}
			}
		}
	}

	return nil, false
}

func assertConst(schema map[string]interface{}, data map[string]interface{}, key string) error {
	expected, ok := constValue(schema, key)
	if !ok {
		return nil
	}
	if value, ok := data[key]; !ok || value != expected {
		return fmt.Errorf("%s does not match const %v", key, expected)
	}
	return nil
}
