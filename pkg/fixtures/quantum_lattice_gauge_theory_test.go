package fixtures

import (
	"context"
	"strings"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestQuantumLatticeGaugeTheoryPipeline(t *testing.T) {
	dag := QuantumLatticeGaugeTheoryPipeline()
	if dag == nil {
		t.Fatal("expected non-nil DAG")
	}
	if dag.TenantID != "high-energy-physics-lab" {
		t.Errorf("expected tenant high-energy-physics-lab, got %s", dag.TenantID)
	}
	if len(dag.Steps) != 4 {
		t.Errorf("expected 4 steps, got %d", len(dag.Steps))
	}

	h1 := &GaugeFieldThermalizationHandler{}
	out1, err := h1.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}
	if !strings.Contains(string(out1.Output), "beta_coupling") {
		t.Errorf("missing beta_coupling")
	}

	h4 := &GradientFlowChargeHandler{}
	out4, err := h4.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 4 failed: %v", err)
	}
	if !strings.Contains(string(out4.Output), "topological_susceptibility_mev") {
		t.Errorf("missing topological_susceptibility_mev")
	}
}
