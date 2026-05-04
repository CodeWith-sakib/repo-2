package core

import (
	"testing"
)

func TestWorkflowSchemaEvolutionValidator(t *testing.T) {
	val := NewWorkflowSchemaEvolutionValidator()

	v1 := WorkflowSchemaDefinition{
		Version: 1,
		Fields: map[string]SchemaFieldDefinition{
			"order_id": {Name: "order_id", Type: "string", Required: true},
			"amount":   {Name: "amount", Type: "float", Required: true},
		},
	}

	v2Valid := WorkflowSchemaDefinition{
		Version: 2,
		Fields: map[string]SchemaFieldDefinition{
			"order_id": {Name: "order_id", Type: "string", Required: true},
			"amount":   {Name: "amount", Type: "float", Required: true},
			"currency": {Name: "currency", Type: "string", Required: false, DefaultValue: "USD"},
		},
	}

	v3Breaking := WorkflowSchemaDefinition{
		Version: 3,
		Fields: map[string]SchemaFieldDefinition{
			"order_id": {Name: "order_id", Type: "string", Required: true},
			"amount":   {Name: "amount", Type: "float", Required: true},
			"api_token": {Name: "api_token", Type: "string", Required: true}, // required without default
		},
	}

	val.RegisterSchema(1, v1)
	val.RegisterSchema(2, v2Valid)
	val.RegisterSchema(3, v3Breaking)

	// v1 -> v2 backward compatibility check (should pass)
	if err := val.ValidateCompatibility(1, 2, CompatibilityBackward); err != nil {
		t.Fatalf("expected v1->v2 compatibility to pass, got: %v", err)
	}

	// v1 -> v3 backward compatibility check (should fail due to required api_token)
	if err := val.ValidateCompatibility(1, 3, CompatibilityBackward); err == nil {
		t.Error("expected v1->v3 to fail backward compatibility, got nil")
	}

	// Payload validation
	validJSON := []byte(`{"order_id":"ord-100","amount":49.99}`)
	if err := val.ValidatePayload(1, validJSON); err != nil {
		t.Errorf("expected payload valid for v1: %v", err)
	}

	invalidJSON := []byte(`{"order_id":"ord-100"}`)
	if err := val.ValidatePayload(1, invalidJSON); err == nil {
		t.Error("expected error for missing amount field in v1 payload, got nil")
	}
}
