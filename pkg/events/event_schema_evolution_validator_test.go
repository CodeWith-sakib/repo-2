package events

import (
	"errors"
	"testing"
)

func TestEventSchemaEvolutionValidator(t *testing.T) {
	validator := NewEventSchemaEvolutionValidator()

	v1 := map[string]SchemaFieldDef{
		"order_id": {Type: "string", Required: true},
		"amount":   {Type: "number", Required: true},
	}
	if err := validator.RegisterSchema("orders", 1, v1); err != nil {
		t.Fatalf("v1 registration failed: %v", err)
	}

	// Valid backward compatible v2 (optional field added)
	v2Valid := map[string]SchemaFieldDef{
		"order_id": {Type: "string", Required: true},
		"amount":   {Type: "number", Required: true},
		"discount": {Type: "number", Required: false},
	}
	if err := validator.RegisterSchema("orders", 2, v2Valid); err != nil {
		t.Fatalf("v2Valid registration failed: %v", err)
	}

	// Incompatible v3 (type alteration)
	v3Invalid := map[string]SchemaFieldDef{
		"order_id": {Type: "int", Required: true}, // changed string -> int
		"amount":   {Type: "number", Required: true},
	}
	err := validator.RegisterSchema("orders", 3, v3Invalid)
	if !errors.Is(err, ErrIncompatibleSchemaEvolution) {
		t.Fatalf("expected ErrIncompatibleSchemaEvolution, got %v", err)
	}
}
