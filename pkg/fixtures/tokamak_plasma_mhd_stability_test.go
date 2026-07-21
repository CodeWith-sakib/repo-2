package fixtures

import (
	"context"
	"strings"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestTokamakPlasmaMHDStabilityPipeline(t *testing.T) {
	dag := TokamakPlasmaMHDStabilityPipeline()
	if dag == nil {
		t.Fatal("expected non-nil DAG")
	}
	if dag.TenantID != "iter-fusion-energy-project" {
		t.Errorf("expected iter-fusion-energy-project, got %s", dag.TenantID)
	}
	if len(dag.Steps) != 4 {
		t.Errorf("expected 4 steps, got %d", len(dag.Steps))
	}

	h1 := &GradShafranovSolverHandler{}
	out1, err := h1.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}
	if !strings.Contains(string(out1.Output), "plasma_current_ma") {
		t.Errorf("missing plasma_current_ma")
	}

	h4 := &RMPCoilDriveHandler{}
	out4, err := h4.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 4 failed: %v", err)
	}
	if !strings.Contains(string(out4.Output), "elm_frequency_suppressed_hz") {
		t.Errorf("missing elm_frequency_suppressed_hz")
	}
}
