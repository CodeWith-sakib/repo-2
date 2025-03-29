package core

import (
	"testing"
)

func TestTemplateMacroExpander(t *testing.T) {
	exp := NewTemplateMacroExpander(true)
	ctx := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
			"id":   42,
		},
		"env": "production",
	}

	tmpl := "Hello {{ user.name }}, user ID is {{ user.id }} in {{ env }}"
	out, err := exp.ExpandString(tmpl, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "Hello Alice, user ID is 42 in production"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}

	_, err = exp.ExpandString("Missing {{ missing.key }}", ctx)
	if err == nil {
		t.Error("expected error for missing key in strict mode")
	}
}
