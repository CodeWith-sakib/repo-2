package fixtures

import (
	"context"
	"strings"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestGeostationaryHallThrusterStationkeepingPipeline(t *testing.T) {
	dag := GeostationaryHallThrusterStationkeepingPipeline()
	if dag == nil {
		t.Fatal("expected non-nil DAG")
	}
	if dag.TenantID != "telecom-satellite-operations" {
		t.Errorf("expected telecom-satellite-operations, got %s", dag.TenantID)
	}
	if len(dag.Steps) != 4 {
		t.Errorf("expected 4 steps, got %d", len(dag.Steps))
	}

	h1 := &LuniSolarPropagatorHandler{}
	out1, err := h1.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}
	if !strings.Contains(string(out1.Output), "delta_inclination_deg_per_year") {
		t.Errorf("missing delta_inclination_deg_per_year")
	}

	h4 := &DeadbandConfinementEvalHandler{}
	out4, err := h4.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 4 failed: %v", err)
	}
	if !strings.Contains(string(out4.Output), "DEADBAND_COMPLIANT") {
		t.Errorf("missing DEADBAND_COMPLIANT")
	}
}
