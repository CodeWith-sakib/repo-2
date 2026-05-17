package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestNanotechSemiconductorLithographyPipeline_Validate(t *testing.T) {
	wf := NanotechSemiconductorLithographyPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestNanotechSemiconductorLithographyPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-fab-01", StepID: "gdsii-oasis-pattern-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"gdsii_pattern_ingest", &GDSIIPatternIngestHandler{}},
		{"euv_source_modeling", &EUVSourceModelingHandler{}},
		{"ilt_opc_solver", &ILTOPCSolverHandler{}},
		{"pellicle_stress_sim", &PellicleStressSimHandler{}},
		{"cd_sem_overlay_align", &CDSEMOverlayAlignHandler{}},
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
