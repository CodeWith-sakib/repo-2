package fixtures

import (
	"context"
	"strings"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestParticleAcceleratorBeamOpticsPipeline(t *testing.T) {
	dag := ParticleAcceleratorBeamOpticsPipeline()
	if dag == nil {
		t.Fatal("expected non-nil DAG")
	}
	if dag.TenantID != "high-energy-physics-lab" {
		t.Errorf("expected high-energy-physics-lab, got %s", dag.TenantID)
	}
	if len(dag.Steps) != 4 {
		t.Errorf("expected 4 steps, got %d", len(dag.Steps))
	}

	h1 := &FODOMatrixSolverHandler{}
	out1, err := h1.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}
	if !strings.Contains(string(out1.Output), "ring_circumference_m") {
		t.Errorf("missing ring_circumference_m")
	}

	h4 := &ChromaticityCorrectionHandler{}
	out4, err := h4.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 4 failed: %v", err)
	}
	if !strings.Contains(string(out4.Output), "head_tail_instability_damped") {
		t.Errorf("missing head_tail_instability_damped")
	}
}
