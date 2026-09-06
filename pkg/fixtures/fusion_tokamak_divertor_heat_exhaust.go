package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// FusionTokamakDivertorHeatExhaustPipeline builds a divertor plasma scrape-off layer (SOL) heat flux mitigation DAG.
func FusionTokamakDivertorHeatExhaustPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_fusion_divertor_heat_exhaust"),
		TenantID:    "iter-fusion-energy-project",
		Name:        "Tokamak Tungsten Monoblock Divertor Scrape-Off Layer Impurity Seeding & Detachment DAG",
		Version:     1,
		Description: "Solves 2D fluid plasma scrape-off layer transport, models neon/nitrogen impurity radiation seeding, tracks localized tungsten divertor monoblock surface temperatures against the 10 MW/m2 technical limit, and establishes stable volumetric plasma detachment.",
		Timeout:     90 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "sol-parallel-heat-transport-solver",
				TaskType: "sol_heat_transport_solver",
			},
			{
				ID:        "neon-impurity-radiation-seeding",
				TaskType:  "impurity_radiation_seeding",
				DependsOn: []string{"sol-parallel-heat-transport-solver"},
			},
			{
				ID:        "divertor-tungsten-monoblock-temp-calc",
				TaskType:  "tungsten_monoblock_temp_calc",
				DependsOn: []string{"neon-impurity-radiation-seeding"},
			},
			{
				ID:        "plasma-detachment-front-eval",
				TaskType:  "detachment_front_eval",
				DependsOn: []string{"divertor-tungsten-monoblock-temp-calc"},
			},
		},
	}
}

// SOLHeatTransportSolverHandler solves classical Spitzer-Harm parallel thermal conduction along magnetic field lines.
type SOLHeatTransportSolverHandler struct{}

func (h *SOLHeatTransportSolverHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"upstream_electron_temp_ev":120.0,"scrape_off_layer_width_mm":1.85,"unmitigated_peak_heat_flux_mw_m2":45.2}`),
	}, nil
}

// ImpurityRadiationSeedingHandler models radiative cooling from seeded extrinsic neon/nitrogen gas puffing.
type ImpurityRadiationSeedingHandler struct{}

func (h *ImpurityRadiationSeedingHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"neon_puff_rate_torr_l_s":14.5,"radiated_power_fraction":0.82,"sol_power_crossing_separatrix_mw":12.4}`),
	}, nil
}

// TungstenMonoblockTempCalcHandler checks peak surface heat flux on ITER-grade tungsten monoblock armour tiles.
type TungstenMonoblockTempCalcHandler struct{}

func (h *TungstenMonoblockTempCalcHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"mitigated_target_heat_flux_mw_m2":6.8,"monoblock_surface_temp_celsius":920.0,"cooling_water_pipe_temp_celsius":185.0,"below_material_limits":true}`),
	}, nil
}

// DetachmentFrontEvalHandler verifies that electron temperature Te at the target plate drops below 5 eV for full detachment.
type DetachmentFrontEvalHandler struct{}

func (h *DetachmentFrontEvalHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"target_electron_temp_ev":2.8,"target_plasma_pressure_drop_ratio":0.15,"divertor_regime":"STABLE_FULL_DETACHMENT"}`),
	}, nil
}
