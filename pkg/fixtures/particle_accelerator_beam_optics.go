package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// ParticleAcceleratorBeamOpticsPipeline builds a synchrotron storage ring particle beam tracking DAG.
func ParticleAcceleratorBeamOpticsPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_synchrotron_beam_optics"),
		TenantID:    "high-energy-physics-lab",
		Name:        "7-GeV Electron Storage Ring Twiss Beta Function & Beam Emittance Matching DAG",
		Version:     1,
		Description: "Models quadrupole focusing and defocusing magnets (FODO cell), computes Courant-Snyder Twiss parameters (beta, alpha, gamma), calculates natural horizontal beam emittance, and optimizes sextupole chromaticity correction.",
		Timeout:     90 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "fodo-lattice-transfer-matrix-solver",
				TaskType: "fodo_matrix_solver",
			},
			{
				ID:        "twiss-beta-dispersion-function-eval",
				TaskType:  "twiss_function_eval",
				DependsOn: []string{"fodo-lattice-transfer-matrix-solver"},
			},
			{
				ID:        "synchrotron-radiation-damping-calc",
				TaskType:  "radiation_damping_calc",
				DependsOn: []string{"twiss-beta-dispersion-function-eval"},
			},
			{
				ID:        "sextupole-chromaticity-correction",
				TaskType:  "chromaticity_correction",
				DependsOn: []string{"synchrotron-radiation-damping-calc"},
			},
		},
	}
}

// FODOMatrixSolverHandler multiplies 6x6 linear transfer matrices across quadrupole and dipole drift spaces.
type FODOMatrixSolverHandler struct{}

func (h *FODOMatrixSolverHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"ring_circumference_m":840.0,"fodo_cell_count":48,"betatron_tune_horizontal":36.22,"betatron_tune_vertical":19.34}`),
	}, nil
}

// TwissFunctionEvalHandler calculates beta-x, beta-y, and horizontal dispersion Dx around the orbit.
type TwissFunctionEvalHandler struct{}

func (h *TwissFunctionEvalHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"max_beta_x_m":24.5,"max_beta_y_m":18.2,"max_dispersion_dx_m":0.35,"beam_stay_clear_verified":true}`),
	}, nil
}

// RadiationDampingCalcHandler calculates energy loss per turn and natural horizontal emittance in nanometer-radians.
type RadiationDampingCalcHandler struct{}

func (h *RadiationDampingCalcHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"energy_loss_per_turn_mev":3.2,"damping_time_horizontal_ms":4.5,"natural_emittance_nm_rad":1.42}`),
	}, nil
}

// ChromaticityCorrectionHandler adjusts harmonic sextupole magnets to set natural chromaticity to positive values.
type ChromaticityCorrectionHandler struct{}

func (h *ChromaticityCorrectionHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"corrected_chromaticity_x":2.0,"corrected_chromaticity_y":2.0,"head_tail_instability_damped":true,"beam_lifetime_hours":14.5}`),
	}, nil
}
