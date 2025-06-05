package postgres

import (
	"context"
	"testing"
)

func TestQueryExplainer(t *testing.T) {
	exp := NewQueryExplainer(nil)
	res, err := exp.Explain(context.Background(), "SELECT * FROM workflows")
	if err != nil || res != "EXPLAIN SELECT * FROM workflows" {
		t.Errorf("unexpected explain result: %v, %v", res, err)
	}
}
