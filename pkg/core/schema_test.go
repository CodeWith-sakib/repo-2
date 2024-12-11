package core

import (
	"testing"
)

func TestJSONSchemaValidation(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"required": ["name", "age"],
		"properties": {
			"name": {"type": "string", "minLength": 2, "pattern": "^[a-zA-Z]+$"},
			"age": {"type": "number", "minimum": 18},
			"tags": {
				"type": "array",
				"items": {"type": "string"}
			}
		}
	}`

	schema, err := CompileSchema([]byte(schemaJSON))
	if err != nil {
		t.Fatalf("failed to compile schema: %v", err)
	}

	// Valid payload
	validPayload := `{"name": "Alice", "age": 30, "tags": ["admin", "dev"]}`
	if err := schema.Validate([]byte(validPayload)); err != nil {
		t.Fatalf("expected valid payload to pass: %v", err)
	}

	// Invalid: missing required field "age"
	missingAge := `{"name": "Bob"}`
	if err := schema.Validate([]byte(missingAge)); err == nil {
		t.Fatal("expected error for missing required field, got nil")
	}

	// Invalid: age < 18
	underage := `{"name": "Charlie", "age": 16}`
	if err := schema.Validate([]byte(underage)); err == nil {
		t.Fatal("expected error for minimum age violation, got nil")
	}

	// Invalid: pattern violation in name
	badPattern := `{"name": "David123", "age": 25}`
	if err := schema.Validate([]byte(badPattern)); err == nil {
		t.Fatal("expected error for pattern violation, got nil")
	}
}
