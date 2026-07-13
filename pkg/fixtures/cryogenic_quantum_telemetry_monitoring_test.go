package fixtures

import (
	"context"
	"strings"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestCryogenicQuantumTelemetryPipeline(t *testing.T) {
	dag := CryogenicQuantumTelemetryPipeline()
	if dag == nil {
		t.Fatal("expected non-nil DAG")
	}
	if dag.TenantID != "quantum-processor-foundry" {
		t.Errorf("expected quantum-processor-foundry, got %s", dag.TenantID)
	}
	if len(dag.Steps) != 3 {
		t.Errorf("expected 3 steps, got %d", len(dag.Steps))
	}

	h1 := &FridgeThermalSensingHandler{}
	out1, err := h1.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}
	if !strings.Contains(string(out1.Output), "mixing_chamber_temp_mk") {
		t.Errorf("missing mixing_chamber_temp_mk")
	}

	h3 := &QubitCoherenceSweepHandler{}
	out3, err := h3.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 3 failed: %v", err)
	}
	if !strings.Contains(string(out3.Output), "t1_relaxation_us") {
		t.Errorf("missing t1_relaxation_us")
	}
}
