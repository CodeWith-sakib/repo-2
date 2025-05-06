package fixtures

import (
	"testing"
)

func TestOrderFulfillmentPipeline(t *testing.T) {
	wf := NewOrderFulfillmentPipeline()
	if len(wf.Steps) != 6 {
		t.Fatalf("expected 6 steps in order fulfillment pipeline, got %d", len(wf.Steps))
	}

	dag, err := wf.BuildDAG()
	if err != nil {
		t.Fatalf("failed building DAG for order fulfillment pipeline: %v", err)
	}

	roots := dag.GetRootNodes()
	if len(roots) != 1 || roots[0] != "validate_order" {
		t.Errorf("expected root step validate_order, got %v", roots)
	}
}
