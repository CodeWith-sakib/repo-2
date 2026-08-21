package fixtures

import (
	"context"
	"strings"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestHypersonicBoundaryLayerTransitionPipeline(t *testing.T) {
	dag := HypersonicBoundaryLayerTransitionPipeline()
	if dag == nil {
		t.Fatal("expected non-nil DAG")
	}
	if dag.TenantID != "aerodynamics-research-institute" {
		t.Errorf("expected aerodynamics-research-institute, got %s", dag.TenantID)
	}
	if len(dag.Steps) != 4 {
		t.Errorf("expected 4 steps, got %d", len(dag.Steps))
	}

	h1 := &CompressibleSimilaritySolverHandler{}
	out1, err := h1.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}
	if !strings.Contains(string(out1.Output), "freestream_mach") {
		t.Errorf("missing freestream_mach")
	}

	h4 := &SkinFrictionEvalHandler{}
	out4, err := h4.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 4 failed: %v", err)
	}
	if !strings.Contains(string(out4.Output), "TURBULENT_EQUILIBRIUM") {
		t.Errorf("missing TURBULENT_EQUILIBRIUM")
	}
}
