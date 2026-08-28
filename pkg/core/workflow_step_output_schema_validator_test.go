package core

import (
	"errors"
	"testing"
)

func TestStepOutputSchemaValidator(t *testing.T) {
	validator := NewStepOutputSchemaValidator()

	validator.RegisterContract("payment_processor", map[string]OutputContractField{
		"transaction_id": {Type: "string", Required: true},
		"amount":         {Type: "number", Required: true},
		"settled":        {Type: "boolean", Required: false},
	})

	validJSON := []byte(`{"transaction_id": "tx_123", "amount": 99.5, "settled": true}`)
	if err := validator.ValidateJSON("payment_processor", validJSON); err != nil {
		t.Fatalf("unexpected error for valid payload: %v", err)
	}

	// Missing required field
	missingJSON := []byte(`{"amount": 99.5}`)
	err := validator.ValidateJSON("payment_processor", missingJSON)
	if !errors.Is(err, ErrOutputSchemaMismatch) {
		t.Fatalf("expected ErrOutputSchemaMismatch, got %v", err)
	}

	// Wrong type
	wrongTypeJSON := []byte(`{"transaction_id": "tx_123", "amount": "invalid_number_string"}`)
	err = validator.ValidateJSON("payment_processor", wrongTypeJSON)
	if !errors.Is(err, ErrOutputSchemaMismatch) {
		t.Fatalf("expected ErrOutputSchemaMismatch for wrong type, got %v", err)
	}
}
