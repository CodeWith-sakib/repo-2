package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestNuclearFusionStellaratorCoilPipeline_Validate(t *testing.T) {
	wf := NuclearFusionStellaratorCoilPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestNuclearFusionStellaratorCoilPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-stellarator-01", StepID: "vmec-mhd-equilibrium-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"vmec_mhd_equilibrium_ingest", &VMECMHDHandler{}},
		{"neoclassical_transport_calc", &NeoclassicalTransportHandler{}},
		{"nonplanar_coil_fourier_opt", &NonplanarCoilFourierOptHandler{}},
		{"regcoil_surface_current_solve", &REGCOILSurfaceCurrentHandler{}},
		{"alpha_particle_loss_eval", &AlphaParticleLossEvalHandler{}},
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
