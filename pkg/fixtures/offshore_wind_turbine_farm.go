package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// OffshoreWindTurbineFarmPipeline builds a SCADA offshore renewable yaw & pitch optimization DAG.
func OffshoreWindTurbineFarmPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_offshore_wind_scada"),
		TenantID:    "renewable-offshore-wind",
		Name:        "15MW Offshore Wind Turbine SCADA Wake Steering & Yaw Pitch DAG",
		Version:     1,
		Description: "Ingests LiDAR nacelle wind vector profiles, models Jensen-Park downwind wake deficit, optimizes cooperative active yaw offset angles to maximize farm-level power production, monitors gearbox planetary bearing acoustics, and transmits subsea HVDC cable inverter setpoints.",
		Timeout:     60 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "nacelle-lidar-wind-vector-ingest",
				TaskType: "lidar_wind_vector_ingest",
			},
			{
				ID:        "jensen-park-wake-deficit-sim",
				TaskType:  "wake_deficit_sim",
				DependsOn: []string{"nacelle-lidar-wind-vector-ingest"},
			},
			{
				ID:        "active-yaw-wake-steering-opt",
				TaskType:  "active_yaw_steering_opt",
				DependsOn: []string{"jensen-park-wake-deficit-sim"},
			},
			{
				ID:        "planetary-gearbox-acoustic-check",
				TaskType:  "gearbox_acoustic_check",
				DependsOn: []string{"nacelle-lidar-wind-vector-ingest"},
			},
			{
				ID:        "hvdc-subsea-inverter-dispatch",
				TaskType:  "hvdc_inverter_dispatch",
				DependsOn: []string{"active-yaw-wake-steering-opt", "planetary-gearbox-acoustic-check"},
			},
		},
	}
}

// LiDARWindVectorHandler parses pulsed laser doppler wind speed and direction vectors.
type LiDARWindVectorHandler struct{}

func (h *LiDARWindVectorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"hub_height_wind_speed_mps":14.2,"wind_direction_deg":245.8,"turbulence_intensity_pct":8.4,"air_density_kg_m3":1.24}`),
	}, nil
}

// WakeDeficitSimHandler models downstream turbulent wake expansion and velocity deficit.
type WakeDeficitSimHandler struct{}

func (h *WakeDeficitSimHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"downwind_turbines_shadowed":6,"mean_wake_deficit_pct":18.5,"rotor_thrust_coefficient":0.78}`),
	}, nil
}

// ActiveYawSteeringHandler computes deliberate misalignment yaw offsets to deflect wake away from downstream turbines.
type ActiveYawSteeringHandler struct{}

func (h *ActiveYawSteeringHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"turbine_01_yaw_offset_deg":4.5,"turbine_02_yaw_offset_deg":2.8,"farm_power_gain_mw":3.8,"net_efficiency_pct":103.2}`),
	}, nil
}

// GearboxAcousticHandler performs high-frequency acoustic emission monitoring for micro-pitting.
type GearboxAcousticHandler struct{}

func (h *GearboxAcousticHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"high_speed_stage_kurtosis":3.1,"bearing_temp_c":62.4,"oil_debris_particle_count":14,"health_state":"NORMAL"}`),
	}, nil
}

// HVDCInverterDispatchHandler transmits real and reactive power setpoints to the offshore substation.
type HVDCInverterDispatchHandler struct{}

func (h *HVDCInverterDispatchHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"total_active_power_mw":148.5,"reactive_power_mvar":12.0,"dc_bus_voltage_kv":320.0,"subsea_link_status":"COMMITTED"}`),
	}, nil
}
