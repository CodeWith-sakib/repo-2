package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// HypersonicBoundaryLayerTransitionPipeline builds a compressible boundary layer laminar-turbulent transition DAG.
func HypersonicBoundaryLayerTransitionPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_cfd_hypersonic_boundary_layer"),
		TenantID:    "aerodynamics-research-institute",
		Name:        "Mach 8 Compressible Boundary Layer Second-Mode Mack Instability & DNS Transition DAG",
		Version:     1,
		Description: "Solves compressible compressible Navier-Stokes similarity equations, computes parabolized stability equations (PSE) N-factor growth for high-frequency trapped acoustic waves, integrates 3D direct numerical simulation (DNS), and predicts laminar-turbulent breakdown location.",
		Timeout:     90 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "compressible-similarity-solver",
				TaskType: "similarity_solver",
			},
			{
				ID:        "parabolized-stability-equations-pse",
				TaskType:  "pse_stability_solver",
				DependsOn: []string{"compressible-similarity-solver"},
			},
			{
				ID:        "mack-mode-nfactor-integration",
				TaskType:  "nfactor_integrator",
				DependsOn: []string{"parabolized-stability-equations-pse"},
			},
			{
				ID:        "laminar-breakdown-skin-friction-eval",
				TaskType:  "skin_friction_eval",
				DependsOn: []string{"mack-mode-nfactor-integration"},
			},
		},
	}
}

// CompressibleSimilaritySolverHandler integrates compressible Illingworth-Stewartson self-similar profile.
type CompressibleSimilaritySolverHandler struct{}

func (h *CompressibleSimilaritySolverHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"freestream_mach":8.0,"wall_temperature_ratio":0.3,"displacement_thickness_mm":1.85,"momentum_thickness_mm":0.42}`),
	}, nil
}

// PSEStabilitySolverHandler solves non-parallel parabolized stability equations for trapped acoustic Mack modes.
type PSEStabilitySolverHandler struct{}

func (h *PSEStabilitySolverHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"second_mode_frequency_khz":320.0,"phase_speed_ratio":0.92,"spatial_growth_rate_per_m":45.2}`),
	}, nil
}

// NFactorIntegratorHandler integrates spatial growth rate alpha-i to calculate envelope e^N transition criteria.
type NFactorIntegratorHandler struct{}

func (h *NFactorIntegratorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"envelope_n_factor":9.2,"critical_n_transition_threshold":9.0,"transition_onset_location_m":0.742}`),
	}, nil
}

// SkinFrictionEvalHandler calculates wall shear stress and local Stanton heat transfer numbers after breakdown.
type SkinFrictionEvalHandler struct{}

func (h *SkinFrictionEvalHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"laminar_cf":0.00045,"turbulent_peak_cf":0.00285,"stanton_number":0.00185,"state":"TURBULENT_EQUILIBRIUM"}`),
	}, nil
}
