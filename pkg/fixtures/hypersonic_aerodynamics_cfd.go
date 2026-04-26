package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// HypersonicAerodynamicsCFDPipeline builds a Mach 7+ scramjet aerothermodynamics CFD simulation DAG.
func HypersonicAerodynamicsCFDPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_hypersonic_cfd_sim"),
		TenantID:    "aerospace-propulsion-lab",
		Name:        "Mach 7 Scramjet Shockwave-Boundary Layer Aerothermodynamics CFD DAG",
		Version:     1,
		Description: "Generates unstructured 3D tetrahedral prism boundary-layer mesh, solves compressible Navier-Stokes equations with Spalart-Allmaras turbulence closure, models non-equilibrium vibrational-dissociation chemistry at 2500K shock front, and calculates Stanton convective heat flux numbers.",
		Timeout:     240 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "cad-geometry-prism-mesh-gen",
				TaskType: "cfd_mesh_generation",
			},
			{
				ID:        "navier-stokes-compressible-solver",
				TaskType:  "navier_stokes_solver",
				DependsOn: []string{"cad-geometry-prism-mesh-gen"},
			},
			{
				ID:        "shock-boundary-layer-interaction",
				TaskType:  "shock_boundary_layer_sim",
				DependsOn: []string{"navier-stokes-compressible-solver"},
			},
			{
				ID:        "chemical-non-equilibrium-dissociation",
				TaskType:  "air_chemistry_dissociation",
				DependsOn: []string{"shock-boundary-layer-interaction"},
			},
			{
				ID:        "stanton-convective-heat-flux-eval",
				TaskType:  "heat_flux_evaluation",
				DependsOn: []string{"chemical-non-equilibrium-dissociation"},
			},
		},
	}
}

// CFDMeshGenerationHandler generates 45-million cell prism inflation layers with y+ < 1.0.
type CFDMeshGenerationHandler struct{}

func (h *CFDMeshGenerationHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"mesh_cells_total":45000000,"prism_layers":28,"y_plus_max":0.85,"inlet_mach":7.2,"status":"MESH_PARTITIONED"}`),
	}, nil
}

// NavierStokesSolverHandler executes second-order Roe upwind flux scheme with multigrid acceleration.
type NavierStokesSolverHandler struct{}

func (h *NavierStokesSolverHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"iterations_converged":4500,"continuity_residual":1.2e-6,"turbulent_viscosity_ratio":14.5,"cfl_number":2.5}`),
	}, nil
}

// ShockBoundaryLayerSimHandler captures oblique shock reflections and recirculation separation bubbles.
type ShockBoundaryLayerSimHandler struct{}

func (h *ShockBoundaryLayerSimHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"separation_bubble_length_mm":42.5,"shock_angle_deg":18.4,"pressure_peak_ratio":8.4,"boundary_layer_thickening_pct":210.0}`),
	}, nil
}

// AirChemistryDissociationHandler models 5-species (N2, O2, NO, N, O) finite-rate Park kinetics.
type AirChemistryDissociationHandler struct{}

func (h *AirChemistryDissociationHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"stagnation_temp_kelvin":2850,"oxygen_dissociation_pct":64.2,"nitric_oxide_mole_fraction":0.048,"chem_equilibrium":false}`),
	}, nil
}

// HeatFluxEvaluationHandler outputs wall surface convective Stanton and Nusselt heat transfer rates.
type HeatFluxEvaluationHandler struct{}

func (h *HeatFluxEvaluationHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"max_surface_heat_flux_mw_m2":4.8,"stanton_number":0.00185,"leading_edge_temp_c":1480.0,"thermal_protection":"TPS_FEASIBLE"}`),
	}, nil
}
