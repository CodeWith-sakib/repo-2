package fixtures

import (
	"context"
	"strings"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestOrbitalDebrisCollisionConjunctionPipeline(t *testing.T) {
	dag := OrbitalDebrisCollisionConjunctionPipeline()
	if dag == nil {
		t.Fatal("expected non-nil DAG")
	}
	if dag.TenantID != "space-traffic-management-agency" {
		t.Errorf("expected space-traffic-management-agency, got %s", dag.TenantID)
	}
	if len(dag.Steps) != 4 {
		t.Errorf("expected 4 steps, got %d", len(dag.Steps))
	}

	h1 := &SGP4OrbitPropagationHandler{}
	out1, err := h1.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}
	if !strings.Contains(string(out1.Output), "relative_velocity_km_s") {
		t.Errorf("missing relative_velocity_km_s")
	}

	h4 := &AvoidanceDeltaVPlannerHandler{}
	out4, err := h4.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 4 failed: %v", err)
	}
	if !strings.Contains(string(out4.Output), "EXECUTE_AVOIDANCE") {
		t.Errorf("missing EXECUTE_AVOIDANCE")
	}
}
