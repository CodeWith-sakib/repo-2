package sql

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestSQLPluginValidationAndExecute(t *testing.T) {
	p := NewSQLPlugin()

	input, _ := json.Marshal(SQLTaskConfig{
		Driver:  "postgres",
		Query:   "UPDATE orders SET status='processed' WHERE status='pending'",
		MaxRows: 100,
	})

	res, err := p.Execute(context.Background(), worker.StepContext{
		Input: input,
	})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(res.Output, &data); err != nil {
		t.Fatalf("unmarshal output failed: %v", err)
	}

	if data["rows_affected"].(float64) != 2 {
		t.Errorf("expected rows_affected 2, got %v", data["rows_affected"])
	}
}
