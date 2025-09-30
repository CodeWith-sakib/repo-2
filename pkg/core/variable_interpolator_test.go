package core

import (
	"encoding/json"
	"testing"
)

func TestVariableInterpolator_SimpleAndNested(t *testing.T) {
	ctx := NewInterpolationContext()
	ctx.Set("tenant", "acme")
	ctx.Set("user", map[string]interface{}{
		"id":   42,
		"name": "Alice",
		"roles": []interface{}{
			"admin", "operator",
		},
	})

	vi := NewVariableInterpolator()

	res, err := vi.InterpolateString("Welcome ${user.name} to ${tenant}!", ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "Welcome Alice to acme!"
	if res != expected {
		t.Errorf("got %q, want %q", res, expected)
	}

	// Array indexing
	resRole, err := vi.InterpolateString("Role: ${user.roles[1]}", ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resRole != "Role: operator" {
		t.Errorf("got %q, want %q", resRole, "Role: operator")
	}
}

func TestVariableInterpolator_DefaultValue(t *testing.T) {
	ctx := NewInterpolationContext()
	vi := NewVariableInterpolator()

	res, err := vi.InterpolateString("Region: ${config.region:-us-east-1}", ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != "Region: us-east-1" {
		t.Errorf("got %q, want %q", res, "Region: us-east-1")
	}
}

func TestVariableInterpolator_MissingVariableError(t *testing.T) {
	ctx := NewInterpolationContext()
	vi := NewVariableInterpolator()

	_, err := vi.InterpolateString("Missing: ${unknown}", ctx)
	if err == nil {
		t.Error("expected error for unknown variable without default")
	}
}

func TestVariableInterpolator_InterpolateJSON(t *testing.T) {
	ctx := NewInterpolationContext()
	ctx.Set("host", "api.internal")
	ctx.Set("port", 8080)

	raw := json.RawMessage(`{"url":"http://${host}:${port}/v1","tags":["env:${env:-prod}"]}`)

	vi := NewVariableInterpolator()
	interpJSON, err := vi.InterpolateJSON(raw, ctx)
	if err != nil {
		t.Fatalf("interpolate JSON failed: %v", err)
	}

	var out map[string]interface{}
	if err := json.Unmarshal(interpJSON, &out); err != nil {
		t.Fatalf("unmarshal result failed: %v", err)
	}

	if out["url"] != "http://api.internal:8080/v1" {
		t.Errorf("unexpected url: %v", out["url"])
	}
}
