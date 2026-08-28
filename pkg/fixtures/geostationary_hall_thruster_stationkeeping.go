package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// GeostationaryHallThrusterStationkeepingPipeline builds an electric propulsion orbital stationkeeping DAG.
func GeostationaryHallThrusterStationkeepingPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_geo_hall_thruster_stationkeeping"),
		TenantID:    "telecom-satellite-operations",
		Name:        "GEO Satellite Xenon Hall-Effect Thruster North-South Stationkeeping DAG",
		Version:     1,
		Description: "Models lunar-solar third-body gravitational perturbations on orbital inclination, calculates optimal daily burn arcs at orbital nodes, computes magnetic field anode electron Hall drift and xenon propellant ionization efficiency, and verifies sub-0.05 degree latitude/longitude deadband constraints.",
		Timeout:     90 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "luni-solar-perturbation-propagator",
				TaskType: "luni_solar_propagator",
			},
			{
				ID:        "hall-thruster-anode-plasma-solver",
				TaskType:  "hall_plasma_solver",
				DependsOn: []string{"luni-solar-perturbation-propagator"},
			},
			{
				ID:        "nodal-burn-arc-optimizer",
				TaskType:  "nodal_burn_optimizer",
				DependsOn: []string{"hall-thruster-anode-plasma-solver"},
			},
			{
				ID:        "box-deadband-confinement-eval",
				TaskType:  "deadband_confinement_eval",
				DependsOn: []string{"nodal-burn-arc-optimizer"},
			},
		},
	}
}

// LuniSolarPropagatorHandler integrates gravitational torque from Moon and Sun to calculate delta inclination.
type LuniSolarPropagatorHandler struct{}

func (h *LuniSolarPropagatorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"delta_inclination_deg_per_year":0.854,"ascending_node_deg":74.5,"required_delta_v_m_s_per_yr":46.2}`),
	}, nil
}

// HallPlasmaSolverHandler computes ExB electron drift, xenon ionization cross-section, and specific impulse Isp.
type HallPlasmaSolverHandler struct{}

func (h *HallPlasmaSolverHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"anode_power_kw":4.5,"specific_impulse_sec":1850.0,"thrust_millinewtons":245.0,"propellant_utilization_ratio":0.94}`),
	}, nil
}

// NodalBurnOptimizerHandler calculates burn start/stop true anomalies around ascending and descending nodes.
type NodalBurnOptimizerHandler struct{}

func (h *NodalBurnOptimizerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"daily_burn_duration_sec":3420.0,"xenon_consumed_grams":78.4,"node_firing_mode":"DUAL_NODE_SPLIT"}`),
	}, nil
}

// DeadbandConfinementEvalHandler verifies orbital box drift remains inside +/- 0.05 deg latitude/longitude envelope.
type DeadbandConfinementEvalHandler struct{}

func (h *DeadbandConfinementEvalHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"max_latitude_deviation_deg":0.038,"max_longitude_deviation_deg":0.041,"stationkeeping_margin_percent":22.5,"status":"DEADBAND_COMPLIANT"}`),
	}, nil
}
