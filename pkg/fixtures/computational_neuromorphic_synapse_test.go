package fixtures

import (
	"context"
	"strings"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestComputationalNeuromorphicSynapsePipeline(t *testing.T) {
	dag := ComputationalNeuromorphicSynapsePipeline()
	if dag == nil {
		t.Fatal("expected non-nil DAG")
	}
	if dag.TenantID != "neuromorphic-computing-lab" {
		t.Errorf("expected neuromorphic-computing-lab, got %s", dag.TenantID)
	}
	if len(dag.Steps) != 4 {
		t.Errorf("expected 4 steps, got %d", len(dag.Steps))
	}

	h1 := &PoissonSpikeTrainHandler{}
	out1, err := h1.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}
	if !strings.Contains(string(out1.Output), "mean_firing_rate_hz") {
		t.Errorf("missing mean_firing_rate_hz")
	}

	h4 := &HomeostaticScalingHandler{}
	out4, err := h4.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 4 failed: %v", err)
	}
	if !strings.Contains(string(out4.Output), "network_stability") {
		t.Errorf("missing network_stability")
	}
}
