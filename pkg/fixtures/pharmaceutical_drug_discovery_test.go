package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestPharmaceuticalDrugDiscoveryPipeline_Validate(t *testing.T) {
	wf := PharmaceuticalDrugDiscoveryPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestPharmaceuticalDrugDiscoveryPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-drug-screen-01", StepID: "target-kinase-pdb-prep"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"target_pdb_prep", &TargetPDBPrepHandler{}},
		{"energy_minimization", &EnergyMinimizationHandler{}},
		{"autodock_docking", &AutoDockDockingHandler{}},
		{"mm_gbsa_scoring", &MMGBSAScoringHandler{}},
		{"admet_screening", &ADMETScreeningHandler{}},
	}

	for _, h := range handlers {
		res, err := h.handler.Execute(ctx, sctx)
		if err != nil {
			t.Fatalf("handler %s failed: %v", h.name, err)
		}
		if len(res.Output) == 0 {
			t.Errorf("expected non-empty output for handler %s", h.name)
		}
	}
}
