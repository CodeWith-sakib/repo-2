package core

import (
	"testing"
)

func TestWorkflowVariableInterpolator(t *testing.T) {
	interp := NewWorkflowVariableInterpolator()

	ctx := map[string]interface{}{
		"inputs": map[string]interface{}{
			"tenant": "enterprise-a",
			"limit":  50,
		},
	}

	tpl := []byte(`{"tenant_id":"${inputs.tenant}","max_items":${inputs.limit}}`)
	res, err := interp.InterpolateJSONRaw(tpl, ctx)
	if err != nil {
		t.Fatalf("unexpected interpolation error: %v", err)
	}

	expected := `{"tenant_id":"enterprise-a","max_items":50}`
	if string(res) != expected {
		t.Errorf("expected '%s', got '%s'", expected, string(res))
	}

	// Missing variable returns error
	badTpl := []byte(`{"id":"${inputs.missing}"}`)
	_, err = interp.InterpolateJSONRaw(badTpl, ctx)
	if err == nil {
		t.Error("expected error on missing variable path, got nil")
	}
}
