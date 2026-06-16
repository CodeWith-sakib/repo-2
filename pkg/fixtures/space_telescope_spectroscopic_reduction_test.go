package fixtures

import (
	"context"
	"strings"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestSpaceTelescopeSpectroscopicReductionPipeline(t *testing.T) {
	dag := SpaceTelescopeSpectroscopicReductionPipeline()
	if dag == nil {
		t.Fatal("expected non-nil DAG")
	}
	if dag.TenantID != "astrophysics-data-center" {
		t.Errorf("expected astrophysics-data-center, got %s", dag.TenantID)
	}
	if len(dag.Steps) != 4 {
		t.Errorf("expected 4 steps, got %d", len(dag.Steps))
	}

	h1 := &DetectorRampFittingHandler{}
	out1, err := h1.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}
	if !strings.Contains(string(out1.Output), "cosmic_ray_jumps_detected") {
		t.Errorf("missing cosmic_ray_jumps_detected")
	}

	h4 := &FluxCalibration1DHandler{}
	out4, err := h4.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 4 failed: %v", err)
	}
	if !strings.Contains(string(out4.Output), "spectroscopic_redshift_z") {
		t.Errorf("missing spectroscopic_redshift_z")
	}
}
