package contracts

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/andreasstove999/ecommerce-system/cart-service-go/internal/cart"
	"github.com/google/uuid"
)

func TestBuildCartCheckedOutEvent(t *testing.T) {
	now := time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC)
	c := &cart.Cart{
		ID:     "a9c9bf1d-32f2-46a0-9243-97c2cf8a6c4a",
		UserID: "1d439ea2-c678-4f2a-9ca9-d8a9755a6a5d",
		Items: []cart.Item{
			{ProductID: "15b50d93-e94b-4e2b-aba8-9ed785a7cdf6", Quantity: 2, Price: 3.5},
		},
		Total: 7.0,
	}

	env := BuildCartCheckedOutEvent(c, EnvelopeOptions{
		PartitionKey:  c.ID,
		Sequence:      42,
		Producer:      CartServiceProducer,
		SchemaPath:    CartCheckedOutEnvelopedSchemaPath,
		CorrelationID: "53b0fd3e-8d6b-49af-8c1f-12cf4182c2f7",
		CausationID:   "63b0fd3e-8d6b-49af-8c1f-12cf4182c2f7",
		EventID:       "73b0fd3e-8d6b-49af-8c1f-12cf4182c2f7",
		OccurredAt:    now,
	})

	if env.EventName != CartCheckedOutEventName {
		t.Fatalf("unexpected event name %s", env.EventName)
	}
	if env.EventVersion != CartCheckedOutEventVersion {
		t.Fatalf("unexpected event version %d", env.EventVersion)
	}
	if env.EventID != "73b0fd3e-8d6b-49af-8c1f-12cf4182c2f7" {
		t.Fatalf("expected provided event id to be used, got %s", env.EventID)
	}
	if env.PartitionKey != c.ID {
		t.Fatalf("expected partition key %s, got %s", c.ID, env.PartitionKey)
	}
	if env.Sequence != 42 {
		t.Fatalf("expected sequence to be 42, got %d", env.Sequence)
	}
	if env.CorrelationID != "53b0fd3e-8d6b-49af-8c1f-12cf4182c2f7" {
		t.Fatalf("unexpected correlation id %s", env.CorrelationID)
	}
	if env.CausationID != "63b0fd3e-8d6b-49af-8c1f-12cf4182c2f7" {
		t.Fatalf("unexpected causation id %s", env.CausationID)
	}
	if env.Schema != CartCheckedOutEnvelopedSchemaPath {
		t.Fatalf("unexpected schema path %s", env.Schema)
	}
	if env.Payload.Timestamp != now {
		t.Fatalf("expected payload timestamp to mirror occurredAt, got %s", env.Payload.Timestamp)
	}
	if len(env.Payload.Items) != 1 || env.Payload.Items[0].ProductID != c.Items[0].ProductID {
		t.Fatalf("payload items not copied correctly: %+v", env.Payload.Items)
	}
}

func TestCartCheckedOutEnvelopeSchemaValidation(t *testing.T) {
	makeEnvelope := func() EventEnvelope {
		now := time.Date(2024, time.March, 1, 12, 0, 0, 0, time.UTC)
		c := &cart.Cart{
			ID:     "e4a9ae13-b4a6-489a-a5fd-89bd88d8e1a5",
			UserID: "f8f98928-87f6-4fce-8c55-7a7a62012f1a",
			Items: []cart.Item{
				{ProductID: "f4007d5d-0212-4bf0-996a-bfde9e0f0170", Quantity: 1, Price: 10.0},
				{ProductID: "bb0a9128-b176-4c0c-9240-8c9a25ffbfc8", Quantity: 2, Price: 5.0},
			},
			Total: 20.0,
		}

		return BuildCartCheckedOutEvent(c, EnvelopeOptions{
			PartitionKey:  c.ID,
			Sequence:      1,
			Producer:      CartServiceProducer,
			SchemaPath:    CartCheckedOutEnvelopedSchemaPath,
			CorrelationID: uuid.NewString(),
			EventID:       uuid.NewString(),
			OccurredAt:    now,
		})
	}

	valid := makeEnvelope()
	assertEnvelopeValid(t, valid)

	t.Run("event name mismatch", func(t *testing.T) {
		invalid := makeEnvelope()
		invalid.EventName = "WrongEvent"
		assertEnvelopeInvalid(t, invalid)
	})

	t.Run("missing partition key", func(t *testing.T) {
		invalid := makeEnvelope()
		invalid.PartitionKey = ""
		assertEnvelopeInvalid(t, invalid)
	})

	t.Run("missing sequence", func(t *testing.T) {
		invalid := makeEnvelope()
		invalid.Sequence = 0
		assertEnvelopeInvalid(t, invalid)
	})

	t.Run("wrong schema path", func(t *testing.T) {
		invalid := makeEnvelope()
		invalid.Schema = "contracts/events/cart/CartCheckedOut.v1.payload.schema.json"
		assertEnvelopeInvalid(t, invalid)
	})

	t.Run("payload missing field", func(t *testing.T) {
		invalid := makeEnvelope()
		invalid.Payload.CartID = ""
		assertEnvelopeInvalid(t, invalid)
	})
}

func assertEnvelopeValid(t *testing.T, env EventEnvelope) {
	t.Helper()
	if err := validateEnvelopeAgainstSchema(env); err != nil {
		t.Fatalf("expected envelope to be valid, got error: %v", err)
	}
}

func assertEnvelopeInvalid(t *testing.T, env EventEnvelope) {
	t.Helper()
	if err := validateEnvelopeAgainstSchema(env); err == nil {
		t.Fatalf("expected envelope to be invalid")
	}
}

func validateEnvelopeAgainstSchema(env EventEnvelope) error {
	if env.PartitionKey == "" {
		return errors.New("partitionKey is required")
	}
	if env.Sequence <= 0 {
		return errors.New("sequence must be positive")
	}

	payload, err := marshalAny(env)
	if err != nil {
		return err
	}

	_, filename, _, _ := runtime.Caller(0)
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", ".."))
	schemaPath := filepath.Join(repoRoot, "contracts", "events", "cart", "CartCheckedOut.v1.enveloped.schema.json")
	schema, baseDir, err := loadSchema(schemaPath)
	if err != nil {
		return err
	}

	return validateNode(schema, payload, baseDir, nil)
}

func marshalAny(v any) (any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	var result any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return result, nil
}

func loadSchema(path string) (map[string]any, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("read schema: %w", err)
	}
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, "", fmt.Errorf("parse schema: %w", err)
	}
	return schema, filepath.Dir(path), nil
}

func validateNode(schema map[string]any, value any, baseDir string, allowedProps map[string]struct{}) error {
	if ref, ok := schema["$ref"].(string); ok {
		resolved, nextBase, err := resolveRef(ref, baseDir)
		if err != nil {
			return err
		}
		return validateNode(resolved, value, nextBase, allowedProps)
	}

	if allOf, ok := schema["allOf"].([]any); ok {
		unionProps, err := collectAllowedProperties(allOf, baseDir)
		if err != nil {
			return err
		}
		for key := range allowedProps {
			unionProps[key] = struct{}{}
		}
		for _, sub := range allOf {
			subSchema, nextBase, err := resolveSchemaNode(sub, baseDir)
			if err != nil {
				return err
			}
			if err := validateNode(subSchema, value, nextBase, unionProps); err != nil {
				return err
			}
		}
	}

	if constVal, ok := schema["const"]; ok {
		if !reflect.DeepEqual(value, constVal) {
			return fmt.Errorf("value %v does not equal const %v", value, constVal)
		}
	}

	if enumVals, ok := schema["enum"].([]any); ok {
		match := false
		for _, v := range enumVals {
			if reflect.DeepEqual(v, value) {
				match = true
				break
			}
		}
		if !match {
			return fmt.Errorf("value %v not in enum %v", value, enumVals)
		}
	}

	if typeVal, ok := schema["type"].(string); ok {
		switch typeVal {
		case "object":
			obj, ok := value.(map[string]any)
			if !ok {
				return fmt.Errorf("expected object, got %T", value)
			}
			if req, ok := schema["required"].([]any); ok {
				for _, field := range req {
					key := field.(string)
					if _, exists := obj[key]; !exists {
						return fmt.Errorf("missing required field %s", key)
					}
				}
			}
			if addl, ok := schema["additionalProperties"].(bool); ok && !addl {
				allowed := make(map[string]struct{})
				for key := range allowedProps {
					allowed[key] = struct{}{}
				}
				if props, ok := schema["properties"].(map[string]any); ok {
					for key := range props {
						allowed[key] = struct{}{}
					}
				}
				for key := range obj {
					if _, exists := allowed[key]; !exists {
						return fmt.Errorf("unexpected property %s", key)
					}
				}
			}
			if props, ok := schema["properties"].(map[string]any); ok {
				for key, propSchema := range props {
					if val, exists := obj[key]; exists {
						childSchema, nextBase, err := resolveSchemaNode(propSchema, baseDir)
						if err != nil {
							return err
						}
						if err := validateNode(childSchema, val, nextBase, nil); err != nil {
							return fmt.Errorf("%s: %w", key, err)
						}
					}
				}
			}
		case "array":
			arr, ok := value.([]any)
			if !ok {
				return fmt.Errorf("expected array, got %T", value)
			}
			if minItems, ok := schema["minItems"].(float64); ok {
				if len(arr) < int(minItems) {
					return fmt.Errorf("expected at least %d items, got %d", int(minItems), len(arr))
				}
			}
			if itemsSchema, ok := schema["items"]; ok {
				childSchema, nextBase, err := resolveSchemaNode(itemsSchema, baseDir)
				if err != nil {
					return err
				}
				for i, item := range arr {
					if err := validateNode(childSchema, item, nextBase, nil); err != nil {
						return fmt.Errorf("items[%d]: %w", i, err)
					}
				}
			}
		case "string":
			str, ok := value.(string)
			if !ok {
				return fmt.Errorf("expected string, got %T", value)
			}
			if minLen, ok := schema["minLength"].(float64); ok {
				if len(str) < int(minLen) {
					return fmt.Errorf("string too short (min %d)", int(minLen))
				}
			}
			if format, ok := schema["format"].(string); ok {
				if err := validateFormat(str, format); err != nil {
					return err
				}
			}
		case "number", "integer":
			num, ok := value.(float64)
			if !ok {
				return fmt.Errorf("expected number, got %T", value)
			}
			if typeVal == "integer" && math.Mod(num, 1) != 0 {
				return fmt.Errorf("expected integer, got %v", num)
			}
			if min, ok := schema["minimum"].(float64); ok {
				if num < min {
					return fmt.Errorf("expected number >= %v", min)
				}
			}
		}
	}

	return nil
}

func resolveSchemaNode(node any, baseDir string) (map[string]any, string, error) {
	if m, ok := node.(map[string]any); ok {
		if ref, ok := m["$ref"].(string); ok {
			return resolveRef(ref, baseDir)
		}
		return m, baseDir, nil
	}
	return nil, baseDir, fmt.Errorf("unsupported schema node %T", node)
}

func resolveRef(ref, baseDir string) (map[string]any, string, error) {
	path := filepath.Clean(filepath.Join(baseDir, ref))
	schema, nextBase, err := loadSchema(path)
	if err != nil {
		return nil, "", err
	}
	return schema, nextBase, nil
}

func collectAllowedProperties(nodes []any, baseDir string) (map[string]struct{}, error) {
	allowed := make(map[string]struct{})
	for _, node := range nodes {
		schema, nextBase, err := resolveSchemaNode(node, baseDir)
		if err != nil {
			return nil, err
		}
		if err := mergeAllowedProperties(schema, allowed, nextBase); err != nil {
			return nil, err
		}
	}
	return allowed, nil
}

func mergeAllowedProperties(schema map[string]any, allowed map[string]struct{}, baseDir string) error {
	if props, ok := schema["properties"].(map[string]any); ok {
		for key := range props {
			allowed[key] = struct{}{}
		}
	}
	if nestedAllOf, ok := schema["allOf"].([]any); ok {
		nestedAllowed, err := collectAllowedProperties(nestedAllOf, baseDir)
		if err != nil {
			return err
		}
		for key := range nestedAllowed {
			allowed[key] = struct{}{}
		}
	}
	return nil
}

func validateFormat(value, format string) error {
	switch format {
	case "uuid":
		if _, err := uuid.Parse(value); err != nil {
			return fmt.Errorf("invalid uuid: %w", err)
		}
	case "date-time":
		if _, err := time.Parse(time.RFC3339, value); err != nil {
			return fmt.Errorf("invalid date-time: %w", err)
		}
	}
	return nil
}
