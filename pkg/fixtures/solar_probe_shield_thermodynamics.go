package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// SolarProbeShieldThermodynamicsPipeline builds a perihelion thermal protection system (TPS) simulation DAG.
func SolarProbeShieldThermodynamicsPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_solar_probe_tps_thermal"),
		TenantID:    "deep-space-exploration-directorate",
		Name:        "Solar Perihelion 9.86 Solar Radii Carbon-Carbon Thermal Protection System DAG",
		Version:     1,
		Description: "Models 475-sun solar irradiance at 0.046 AU perihelion, calculates Stefan-Boltzmann radiative equilibrium on front alumina-coated carbon foam heatshield, evaluates 1D unsteady Fourier heat conduction across carbon-carbon substrate, and verifies payload umbilical cold-finger temp stays below 30 deg C.",
		Timeout:     90 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "perihelion-solar-flux-irradiance",
				TaskType: "solar_flux_calc",
			},
			{
				ID:        "alumina-coating-re-radiation-solver",
				TaskType:  "re_radiation_solver",
				DependsOn: []string{"perihelion-solar-flux-irradiance"},
			},
			{
				ID:        "carbon-foam-conduction-solver",
				TaskType:  "foam_conduction_solver",
				DependsOn: []string{"alumina-coating-re-radiation-solver"},
			},
			{
				ID:        "umbilical-cold-finger-temp-eval",
				TaskType:  "cold_finger_temp_eval",
				DependsOn: []string{"carbon-foam-conduction-solver"},
			},
		},
	}
}

// SolarFluxCalcHandler calculates solar constant scaling by inverse square law at close perihelion.
type SolarFluxCalcHandler struct{}

func (h *SolarFluxCalcHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"heliocentric_distance_au":0.046,"solar_irradiance_kw_m2":645.0,"sun_angular_diameter_deg":12.5}`),
	}, nil
}

// ReRadiationSolverHandler solves Stefan-Boltzmann T^4 surface equilibrium with diffuse alumina reflectance.
type ReRadiationSolverHandler struct{}

func (h *ReRadiationSolverHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"shield_front_temp_celsius":1370.0,"emissivity_ir":0.88,"solar_absorptance":0.18,"re_radiated_fraction":0.965}`),
	}, nil
}

// FoamConductionSolverHandler solves 1D transient heat conduction through 11.4cm carbon-composite foam core.
type FoamConductionSolverHandler struct{}

func (h *FoamConductionSolverHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"shield_back_temp_celsius":312.0,"thermal_gradient_c_per_cm":92.8,"foam_density_g_cm3":0.16}`),
	}, nil
}

// ColdFingerTempEvalHandler checks instrument bus radiator temperatures in shadow of heatshield.
type ColdFingerTempEvalHandler struct{}

func (h *ColdFingerTempEvalHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"payload_chassis_temp_celsius":26.4,"radiator_temp_celsius":-12.0,"thermal_margin_celsius":23.6,"status":"THERMAL_SURVIVAL_ASSURED"}`),
	}, nil
}
