package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// TokamakPlasmaMHDStabilityPipeline builds a nuclear fusion tokamak magnetohydrodynamics (MHD) equilibrium DAG.
func TokamakPlasmaMHDStabilityPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_tokamak_mhd_stability"),
		TenantID:    "iter-fusion-energy-project",
		Name:        "Tokamak D-T Burning Plasma Magnetohydrodynamic Equilibrium & Sawtooth Crash DAG",
		Version:     1,
		Description: "Solves 2D Grad-Shafranov equilibrium across poloidal magnetic flux surfaces, models safety factor q profile evolution, evaluates Mercier ideal ballooning instability criteria, and triggers resonant magnetic perturbation (RMP) coil ELM suppression.",
		Timeout:     90 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "grad-shafranov-flux-solver",
				TaskType: "grad_shafranov_solver",
			},
			{
				ID:        "safety-factor-q-profile-eval",
				TaskType:  "q_profile_eval",
				DependsOn: []string{"grad-shafranov-flux-solver"},
			},
			{
				ID:        "mercier-ballooning-instability-check",
				TaskType:  "ballooning_check",
				DependsOn: []string{"safety-factor-q-profile-eval"},
			},
			{
				ID:        "rmp-coil-elm-suppression-drive",
				TaskType:  "rmp_coil_drive",
				DependsOn: []string{"mercier-ballooning-instability-check"},
			},
		},
	}
}

// GradShafranovSolverHandler solves non-linear elliptic partial differential equation for poloidal magnetic flux.
type GradShafranovSolverHandler struct{}

func (h *GradShafranovSolverHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"plasma_current_ma":15.0,"toroidal_magnetic_field_tesla":5.3,"plasma_beta_percent":3.8,"poloidal_flux_weber":124.5}`),
	}, nil
}

// QProfileEvalHandler calculates pitch angle of magnetic field lines and locates q=1, q=2 rational surfaces.
type QProfileEvalHandler struct{}

func (h *QProfileEvalHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"q_zero_core":0.98,"q_edge_boundary":3.14,"sawtooth_inversion_radius_m":0.45,"sawtooth_period_ms":85}`),
	}, nil
}

// BallooningCheckHandler verifies localized high-n pressure-driven interchange modes.
type BallooningCheckHandler struct{}

func (h *BallooningCheckHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"mercier_index":0.12,"ballooning_growth_rate_khz":0.0,"edge_localized_mode_risk":"CRITICAL"}`),
	}, nil
}

// RMPCoilDriveHandler applies n=3 3D resonant magnetic perturbation coil fields to suppress ELM bursts.
type RMPCoilDriveHandler struct{}

func (h *RMPCoilDriveHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"rmp_current_kat":45.0,"elm_frequency_suppressed_hz":120.0,"island_width_cm":3.2,"plasma_confinement_h98":1.02}`),
	}, nil
}
