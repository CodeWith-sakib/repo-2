package transform

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/plugin"
)

func TestTransformPluginExecute(t *testing.T) {
	p := NewTransformPlugin()

	cfg, _ := json.Marshal(map[string]interface{}{
		"expressions": map[string]string{
			"userName": "user.name",
			"status":   "data.items[0]",
		},
	})

	input, _ := json.Marshal(map[string]interface{}{
		"user": map[string]interface{}{"name": "Kestrel"},
		"data": map[string]interface{}{
			"items": []interface{}{"active", "pending"},
		},
	})

	resp, err := p.Execute(context.Background(), plugin.ExecutionRequest{
		Config: cfg,
		Input:  input,
	})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	var res map[string]interface{}
	if err := json.Unmarshal(resp.Output, &res); err != nil {
		t.Fatalf("unmarshal output failed: %v", err)
	}

	if res["userName"] != "Kestrel" {
		t.Errorf("expected userName Kestrel, got %v", res["userName"])
	}
	if res["status"] != "active" {
		t.Errorf("expected status active, got %v", res["status"])
	}
}
