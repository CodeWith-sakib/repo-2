package http

import (
	"encoding/json"
	"testing"
)

func TestOpenAPIGenerator(t *testing.T) {
	gen := NewOpenAPIGenerator("KestrelFlow API", "v1.0.0")
	bytes, err := gen.ToJSON()
	if err != nil {
		t.Fatalf("failed generating OpenAPI JSON: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(bytes, &raw); err != nil {
		t.Fatalf("invalid generated JSON: %v", err)
	}

	if raw["openapi"] != "3.0.3" {
		t.Errorf("expected openapi 3.0.3, got %v", raw["openapi"])
	}

	paths := raw["paths"].(map[string]interface{})
	if _, ok := paths["/api/v1/workflows"]; !ok {
		t.Error("expected /api/v1/workflows route in OpenAPI spec")
	}
}
