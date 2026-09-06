package fixtures

import (
	"context"
	"strings"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestFusionTokamakDivertorHeatExhaustPipeline(t *testing.T) {
	dag := FusionTokamakDivertorHeatExhaustPipeline()
	if dag == nil {
		t.Fatal("expected non-nil DAG")
	}
	if dag.TenantID != "iter-fusion-energy-project" {
		t.Errorf("expected iter-fusion-energy-project, got %s", dag.TenantID)
	}
	if len(dag.Steps) != 4 {
		t.Errorf("expected 4 steps, got %d", len(dag.Steps))
	}

	h1 := &SOLHeatTransportSolverHandler{}
	out1, err := h1.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}
	if !strings.Contains(string(out1.Output), "unmitigated_peak_heat_flux_mw_m2") {
		t.Errorf("missing unmitigated_peak_heat_flux_mw_m2")
	}

	h4 := &DetachmentFrontEvalHandler{}
	out4, err := h4.Execute(context.Background(), worker.StepContext{})
	if err != nil {
		t.Fatalf("step 4 failed: %v", err)
	}
	if !strings.Contains(string(out4.Output), "STABLE_FULL_DETACHMENT") {
		t.Errorf("missing STABLE_FULL_DETACHMENT")
	}
}
