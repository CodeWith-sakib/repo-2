package fixtures

import (
	"context"
	"strings"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestHypersonicReentryAerothermodynamicsPipeline(t *testing.T) {
	dag := HypersonicReentryAerothermodynamicsPipeline()
	if dag == nil {
		t.Fatal("expected non-nil DAG")
	}
	if dag.TenantID != "aerospace-hypersonics-division" {
		t.Errorf("expected aerospace-hypersonics-division, got %s", dag.TenantID)
	}
	if len(dag.Steps) != 4 {
		t.Errorf("expected 4 steps, got %d", len(dag.Steps))
	}

	h1 := &BowShockSolverHandler{}
	out1, err := h1.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}
	if !strings.Contains(string(out1.Output), "mach_number") {
		t.Errorf("missing mach_number")
	}

	h4 := &AblativePyrolysisSolverHandler{}
	out4, err := h4.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 4 failed: %v", err)
	}
	if !strings.Contains(string(out4.Output), "THERMAL_INTEGRITY_VERIFIED") {
		t.Errorf("missing THERMAL_INTEGRITY_VERIFIED")
	}
}
