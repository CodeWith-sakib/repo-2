package fixtures

import (
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestHealthcareHL7IngestionPipeline(t *testing.T) {
	wf := NewHealthcareHL7IngestionPipeline()
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
	dag, err := core.BuildDAG(wf.Steps)
	if err != nil {
		t.Fatalf("failed building DAG: %v", err)
	}
	roots := dag.RootNodes()
	if len(roots) != 1 || roots[0] != "receive_mllp_packet" {
		t.Errorf("expected root receive_mllp_packet, got %v", roots)
	}
}
