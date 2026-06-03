package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// GeothermalReservoirHydrodynamicsPipeline builds an enhanced geothermal system (EGS) reservoir model DAG.
func GeothermalReservoirHydrodynamicsPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_geothermal_hydrodynamics"),
		TenantID:    "geothermal-energy-consortium",
		Name:        "Enhanced Geothermal System (EGS) Coupled Hydro-Thermal-Mechanical Reservoir DAG",
		Version:     1,
		Description: "Models crystalline basement rock hydraulic fracture network propagation, fluid injection pressure, high-enthalpy thermal drawdown, and supercritical CO2 working fluid heat extraction lifetime.",
		Timeout:     90 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "acoustic-wellbore-televiewer",
				TaskType: "televiewer_fracture_scan",
			},
			{
				ID:        "hydraulic-shear-stimulation",
				TaskType:  "hydraulic_shear_stimulation",
				DependsOn: []string{"acoustic-wellbore-televiewer"},
			},
			{
				ID:        "thermo-hydraulic-flow-solver",
				TaskType:  "thermo_hydraulic_solver",
				DependsOn: []string{"hydraulic-shear-stimulation"},
			},
		},
	}
}

// TeleviewerFractureScanHandler processes ultrasonic borehole televiewer scans.
type TeleviewerFractureScanHandler struct{}

func (h *TeleviewerFractureScanHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"depth_meters":3500.0,"fracture_density_per_m":2.5,"formation_stress_azimuth_deg":135.0}`),
	}, nil
}

// HydraulicShearStimulationHandler models hydro-shearing along natural joints.
type HydraulicShearStimulationHandler struct{}

func (h *HydraulicShearStimulationHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"injection_pressure_mpa":65.4,"microseismic_cloud_volume_km3":0.85,"permeability_enhancement_ratio":42.0}`),
	}, nil
}

// ThermoHydraulicSolverHandler solves non-isothermal Navier-Stokes and rock thermal conduction.
type ThermoHydraulicSolverHandler struct{}

func (h *ThermoHydraulicSolverHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"production_temp_celsius":215.8,"mass_flow_rate_kgs":75.0,"thermal_drawdown_rate_per_yr":0.008,"net_power_output_mwe":18.4}`),
	}, nil
}
