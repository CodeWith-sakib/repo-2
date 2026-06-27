package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// QuantumLatticeGaugeTheoryPipeline builds a 4D SU(3) lattice quantum chromodynamics (LQCD) Monte Carlo DAG.
func QuantumLatticeGaugeTheoryPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_lqcd_su3_gauge"),
		TenantID:    "high-energy-physics-lab",
		Name:        "4D Euclidean SU(3) Lattice Gauge Theory Hybrid Monte Carlo (HMC) DAG",
		Version:     1,
		Description: "Simulates Wilson plaquette gauge action on a 32^3 x 64 space-time lattice, integrates molecular dynamics Hamilton equations, updates pseudo-fermion clover fields, and calculates topological susceptibility and hadron mass spectrum.",
		Timeout:     120 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "lattice-gauge-field-thermalization",
				TaskType: "gauge_field_thermalization",
			},
			{
				ID:        "hybrid-monte-carlo-trajectory",
				TaskType:  "hmc_trajectory_integration",
				DependsOn: []string{"lattice-gauge-field-thermalization"},
			},
			{
				ID:        "wilson-plaquette-measurement",
				TaskType:  "plaquette_flux_measurement",
				DependsOn: []string{"hybrid-monte-carlo-trajectory"},
			},
			{
				ID:        "topological-charge-gradient-flow",
				TaskType:  "gradient_flow_charge",
				DependsOn: []string{"wilson-plaquette-measurement"},
			},
		},
	}
}

// GaugeFieldThermalizationHandler initializes gauge links with hot/cold start and sweeps heatbath.
type GaugeFieldThermalizationHandler struct{}

func (h *GaugeFieldThermalizationHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"lattice_geometry":"32x32x32x64","beta_coupling":6.0,"thermalization_sweeps":500,"acceptance_rate":0.82}`),
	}, nil
}

// HMCTrajectoryIntegrationHandler runs leapfrog/Omelyan integrator for pseudo-fermion molecular dynamics.
type HMCTrajectoryIntegrationHandler struct{}

func (h *HMCTrajectoryIntegrationHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"trajectory_length":1.0,"time_steps":20,"hamiltonian_delta_h":0.042,"metropolis_accepted":true}`),
	}, nil
}

// PlaquetteFluxMeasurementHandler calculates the average trace of 1x1 Wilson loops across spatial and temporal planes.
type PlaquetteFluxMeasurementHandler struct{}

func (h *PlaquetteFluxMeasurementHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"mean_plaquette":0.5937,"spatial_plaquette":0.5932,"temporal_plaquette":0.5942}`),
	}, nil
}

// GradientFlowChargeHandler applies Wilson flow smoothing to extract topological charge Q and string tension.
type GradientFlowChargeHandler struct{}

func (h *GradientFlowChargeHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"flow_time_t":2.5,"topological_charge_q":-1.004,"topological_susceptibility_mev":182.4,"pion_mass_mev":138.2}`),
	}, nil
}
