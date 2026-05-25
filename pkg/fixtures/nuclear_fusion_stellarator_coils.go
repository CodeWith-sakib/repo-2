package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// NuclearFusionStellaratorCoilPipeline builds an optimized 3D non-planar magnetic stellarator confinement DAG.
func NuclearFusionStellaratorCoilPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_fusion_stellarator_opt"),
		TenantID:    "plasma-physics-fusion-lab",
		Name:        "Wendelstein 7-X Class 3D Quasi-Isodynamic Non-Planar Stellarator Coil Optimization DAG",
		Version:     1,
		Description: "Ingests VMEC 3D magnetohydrodynamic equilibrium flux surfaces, calculates neoclassical transport ripple diffusion coefficients, optimizes 50 non-planar modular superconducting coil Fourier coefficients, solves REGCOIL magnetic surface current distribution, and simulates fast alpha particle collisionless confinement loss fractions.",
		Timeout:     240 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "vmec-mhd-equilibrium-ingest",
				TaskType: "vmec_mhd_equilibrium_ingest",
			},
			{
				ID:        "neoclassical-ripple-transport-calc",
				TaskType:  "neoclassical_transport_calc",
				DependsOn: []string{"vmec-mhd-equilibrium-ingest"},
			},
			{
				ID:        "modular-nonplanar-coil-fourier-opt",
				TaskType:  "nonplanar_coil_fourier_opt",
				DependsOn: []string{"neoclassical-ripple-transport-calc"},
			},
			{
				ID:        "regcoil-surface-current-solve",
				TaskType:  "regcoil_surface_current_solve",
				DependsOn: []string{"modular-nonplanar-coil-fourier-opt"},
			},
			{
				ID:        "fast-alpha-collisionless-loss-eval",
				TaskType:  "alpha_particle_loss_eval",
				DependsOn: []string{"regcoil-surface-current-solve"},
			},
		},
	}
}

// VMECMHDHandler parses Variational Moments Equilibrium Code toroidal flux functions.
type VMECMHDHandler struct{}

func (h *VMECMHDHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"plasma_volume_m3":32.4,"aspect_ratio":10.5,"rotational_transform_iota":0.86,"beta_volume_pct":4.8}`),
	}, nil
}

// NeoclassicalTransportHandler computes monoenergetic drift kinetic equation (DKES) transport matrices.
type NeoclassicalTransportHandler struct{}

func (h *NeoclassicalTransportHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"effective_ripple_epsilon_eff":0.0084,"radial_electric_field_er_kv_m":-14.2,"bootstrap_current_ka":12.5}`),
	}, nil
}

// NonplanarCoilFourierOptHandler minimizes normal magnetic field deviation B_dot_n across winding surface.
type NonplanarCoilFourierOptHandler struct{}

func (h *NonplanarCoilFourierOptHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"modular_coils_count":50,"max_coil_curvature_m_inv":1.85,"coil_to_coil_distance_min_m":0.18,"chi2_b_norm":1.4e-5}`),
	}, nil
}

// REGCOILSurfaceCurrentHandler inverts Laplace equation for continuous current potential on coil winding surface.
type REGCOILSurfaceCurrentHandler struct{}

func (h *REGCOILSurfaceCurrentHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"max_surface_current_density_ma_m":8.9,"tikhonov_regularization_alpha":1e-14,"coil_complexity_score":0.92}`),
	}, nil
}

// AlphaParticleLossEvalHandler tracks 100,000 fusion alpha orbits over 100,000 toroidal transits.
type AlphaParticleLossEvalHandler struct{}

func (h *AlphaParticleLossEvalHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"alpha_particles_simulated":100000,"energy_loss_fraction_pct":2.14,"collisionless_prompt_loss_pct":0.82,"confinement_quality":"REACTOR_GRADE"}`),
	}, nil
}
