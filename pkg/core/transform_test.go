package core

import (
	"encoding/json"
	"testing"
)

func TestPayloadTransformer(t *testing.T) {
	inputJSON := `{
		"user": {
			"profile": {
				"name": "Jane Doe",
				"contacts": ["jane@example.com", "555-1234"]
			}
		},
		"order": {
			"total": 99.95
		}
	}`

	rules := []TransformationRule{
		{SourcePath: "$.user.profile.name", TargetPath: "customer.full_name"},
		{SourcePath: "$.user.profile.contacts[0]", TargetPath: "customer.primary_email"},
		{SourcePath: "$.order.total", TargetPath: "billing.amount"},
		{SourcePath: "$.missing.value", TargetPath: "fallback.status", DefaultVal: "DEFAULT_OK"},
	}

	transformer := NewPayloadTransformer(rules)
	outputBytes, err := transformer.Transform([]byte(inputJSON))
	if err != nil {
		t.Fatalf("transformation failed: %v", err)
	}

	var output map[string]interface{}
	_ = json.Unmarshal(outputBytes, &output)

	name, _ := ExtractJSONPath(output, "$.customer.full_name")
	if name != "Jane Doe" {
		t.Errorf("expected Jane Doe, got %v", name)
	}

	email, _ := ExtractJSONPath(output, "$.customer.primary_email")
	if email != "jane@example.com" {
		t.Errorf("expected jane@example.com, got %v", email)
	}

	fallback, _ := ExtractJSONPath(output, "$.fallback.status")
	if fallback != "DEFAULT_OK" {
		t.Errorf("expected DEFAULT_OK, got %v", fallback)
	}
}
