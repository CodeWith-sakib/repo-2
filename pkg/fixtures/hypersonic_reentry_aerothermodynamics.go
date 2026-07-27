package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// HypersonicReentryAerothermodynamicsPipeline builds a blunt-body atmospheric entry aerothermal DAG.
func HypersonicReentryAerothermodynamicsPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_hypersonic_reentry_aero"),
		TenantID:    "aerospace-hypersonics-division",
		Name:        "Mach 25 Blunt-Body Reentry Plasma Sheath & Non-Equilibrium Aerothermodynamics DAG",
		Version:     1,
		Description: "Models hypersonic shock standoff distance, 5-species (N2, O2, NO, N, O) chemical non-equilibrium dissociation, radiative and convective heat flux via Fay-Riddell formulation, and carbon-phenolic ablative heatshield pyrolysis.",
		Timeout:     90 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "bow-shock-standoff-solver",
				TaskType: "bow_shock_solver",
			},
			{
				ID:        "chemical-nonequilibrium-kinetics",
				TaskType:  "chemical_kinetics_solver",
				DependsOn: []string{"bow-shock-standoff-solver"},
			},
			{
				ID:        "fay-riddell-stagnation-heatflux",
				TaskType:  "stagnation_heatflux_solver",
				DependsOn: []string{"chemical-nonequilibrium-kinetics"},
			},
			{
				ID:        "ablative-pyrolysis-heatshield-solver",
				TaskType:  "ablative_pyrolysis_solver",
				DependsOn: []string{"fay-riddell-stagnation-heatflux"},
			},
		},
	}
}

// BowShockSolverHandler calculates bow shock curvature and detached standoff distance.
type BowShockSolverHandler struct{}

func (h *BowShockSolverHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"mach_number":24.8,"altitude_km":68.5,"shock_standoff_distance_mm":42.5,"post_shock_temp_kelvin":11500.0}`),
	}, nil
}

// ChemicalKineticsSolverHandler integrates Park 1993 chemical reaction rates for high-temp air.
type ChemicalKineticsSolverHandler struct{}

func (h *ChemicalKineticsSolverHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"o2_dissociation_fraction":0.99,"n2_dissociation_fraction":0.48,"electron_number_density_per_m3":4.5e19,"rf_blackout_predicted":true}`),
	}, nil
}

// StagnationHeatfluxSolverHandler calculates catalytic wall convective and radiative flux.
type StagnationHeatfluxSolverHandler struct{}

func (h *StagnationHeatfluxSolverHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"convective_heat_flux_mw_m2":14.2,"radiative_heat_flux_mw_m2":3.8,"total_stagnation_heat_flux_mw_m2":18.0}`),
	}, nil
}

// AblativePyrolysisSolverHandler tracks virgin resin decomposition, char layer recession, and surface gas blowing.
type AblativePyrolysisSolverHandler struct{}

func (h *AblativePyrolysisSolverHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"char_layer_thickness_mm":18.4,"pyrolysis_gas_mass_flux_kg_m2_s":0.125,"heatshield_backface_temp_celsius":145.0,"status":"THERMAL_INTEGRITY_VERIFIED"}`),
	}, nil
}
