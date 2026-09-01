package fixtures

import (
	"context"
	"strings"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestSolarProbeShieldThermodynamicsPipeline(t *testing.T) {
	dag := SolarProbeShieldThermodynamicsPipeline()
	if dag == nil {
		t.Fatal("expected non-nil DAG")
	}
	if dag.TenantID != "deep-space-exploration-directorate" {
		t.Errorf("expected deep-space-exploration-directorate, got %s", dag.TenantID)
	}
	if len(dag.Steps) != 4 {
		t.Errorf("expected 4 steps, got %d", len(dag.Steps))
	}

	h1 := &SolarFluxCalcHandler{}
	out1, err := h1.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}
	if !strings.Contains(string(out1.Output), "solar_irradiance_kw_m2") {
		t.Errorf("missing solar_irradiance_kw_m2")
	}

	h4 := &ColdFingerTempEvalHandler{}
	out4, err := h4.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 4 failed: %v", err)
	}
	if !strings.Contains(string(out4.Output), "THERMAL_SURVIVAL_ASSURED") {
		t.Errorf("missing THERMAL_SURVIVAL_ASSURED")
	}
}
