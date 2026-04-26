package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestHypersonicAerodynamicsCFDPipeline_Validate(t *testing.T) {
	wf := HypersonicAerodynamicsCFDPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestHypersonicAerodynamicsCFDPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-cfd-01", StepID: "cad-geometry-prism-mesh-gen"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"cfd_mesh_generation", &CFDMeshGenerationHandler{}},
		{"navier_stokes_solver", &NavierStokesSolverHandler{}},
		{"shock_boundary_layer_sim", &ShockBoundaryLayerSimHandler{}},
		{"air_chemistry_dissociation", &AirChemistryDissociationHandler{}},
		{"heat_flux_evaluation", &HeatFluxEvaluationHandler{}},
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
