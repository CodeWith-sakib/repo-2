package fixtures

import (
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestFintechPaymentClearingPipeline(t *testing.T) {
	wf := NewFintechPaymentClearingPipeline()
	if len(wf.Steps) != 8 {
		t.Fatalf("expected 8 steps, got %d", len(wf.Steps))
	}
	dag, err := core.BuildDAG(wf.Steps)
	if err != nil {
		t.Fatalf("failed building DAG: %v", err)
	}
	roots := dag.RootNodes()
	if len(roots) != 1 || roots[0] != "parse_iso20022_message" {
		t.Errorf("expected root parse_iso20022_message, got %v", roots)
	}
}
