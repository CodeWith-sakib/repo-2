package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// EnergyGridLoadDispatchPipeline builds a smart power grid economic dispatch DAG.
func EnergyGridLoadDispatchPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_energy_grid_dispatch"),
		TenantID:    "national-grid-iso",
		Name:        "Smart Power Grid 15-Minute Economic Dispatch & Renewable Balancing DAG",
		Version:     1,
		Description: "Ingests real-time PMU phasor measurements, forecasts solar/wind generation, solves non-linear AC optimal power flow (ACOPF), checks N-1 transmission line contingency, and dispatches battery energy storage systems (BESS).",
		Timeout:     60 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "pmu-phasor-ingest",
				TaskType: "pmu_ingest",
			},
			{
				ID:        "renewable-forecast-solar-wind",
				TaskType:  "renewable_forecast",
				DependsOn: []string{"pmu-phasor-ingest"},
			},
			{
				ID:        "acopf-optimal-power-flow",
				TaskType:  "acopf_solver",
				DependsOn: []string{"renewable-forecast-solar-wind"},
			},
			{
				ID:        "n1-contingency-security-analysis",
				TaskType:  "n1_security_analysis",
				DependsOn: []string{"acopf-optimal-power-flow"},
			},
			{
				ID:        "bess-battery-dispatch-setpoint",
				TaskType:  "bess_dispatch",
				DependsOn: []string{"n1-contingency-security-analysis"},
			},
		},
	}
}

// PMUPhasorHandler reads high-frequency synchrophasor measurements.
type PMUPhasorHandler struct{}

func (h *PMUPhasorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"substation_id":"SUB-400KV-NORTH","voltage_magnitude_kv":401.2,"phase_angle_deg":-12.4,"frequency_hz":59.998,"rocof_hz_s":-0.002}`),
	}, nil
}

// GridRenewableForecastHandler models cloud cover and wind speed trajectories.
type GridRenewableForecastHandler struct{}

func (h *GridRenewableForecastHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"solar_mw_expected":450.0,"wind_mw_expected":1200.0,"ramp_rate_mw_min":-15.2,"forecast_horizon_min":15}`),
	}, nil
}

// ACOPFHandler computes least-cost generator setpoints under thermal and voltage constraints.
type ACOPFHandler struct{}

func (h *ACOPFHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"marginal_clearing_price_mwh":38.45,"total_generation_mw":4210.0,"locational_marginal_price_spread":2.10,"converged":true}`),
	}, nil
}

// N1SecurityAnalysisHandler runs contingency screening for unexpected single line trips.
type N1SecurityAnalysisHandler struct{}

func (h *N1SecurityAnalysisHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"contingencies_simulated":142,"overloaded_branches":0,"worst_case_loading_pct":88.4,"security_status":"SECURE"}`),
	}, nil
}

// BESSDispatchHandler sends charge/discharge setpoints to utility battery banks.
type BESSDispatchHandler struct{}

func (h *BESSDispatchHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"bess_units_dispatched":12,"total_power_mw":150.0,"mode":"FAST_FREQUENCY_RESPONSE","roundtrip_efficiency_pct":89.5}`),
	}, nil
}
