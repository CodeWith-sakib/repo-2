package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// GeospatialWildfireSpreadPipeline builds an emergency disaster response & fire simulation DAG.
func GeospatialWildfireSpreadPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_wildfire_spread_sim"),
		TenantID:    "emergency-disaster-management",
		Name:        "Geospatial Wildfire Propagation & Community Evacuation DAG",
		Version:     1,
		Description: "Ingests MODIS/VIIRS thermal satellite anomalies, integrates high-resolution DEM slope contours, calculates Rothermel surface fire rate of spread (ROS), executes Monte Carlo ember spotting simulation, and broadcasts evacuation zone orders via IPAWS CAP alerts.",
		Timeout:     90 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "modis-viirs-satellite-ingest",
				TaskType: "satellite_thermal_ingest",
			},
			{
				ID:        "dem-slope-fuel-layer-prep",
				TaskType:  "dem_fuel_layer_prep",
				DependsOn: []string{"modis-viirs-satellite-ingest"},
			},
			{
				ID:        "rothermel-rate-of-spread-calc",
				TaskType:  "rothermel_ros_calc",
				DependsOn: []string{"dem-slope-fuel-layer-prep"},
			},
			{
				ID:        "monte-carlo-ember-spotting-sim",
				TaskType:  "ember_spotting_sim",
				DependsOn: []string{"rothermel-rate-of-spread-calc"},
			},
			{
				ID:        "ipaws-cap-evacuation-broadcast",
				TaskType:  "ipaws_evacuation_broadcast",
				DependsOn: []string{"monte-carlo-ember-spotting-sim"},
			},
		},
	}
}

// SatelliteThermalIngestHandler ingests VIIRS 375m active fire detection pixels.
type SatelliteThermalIngestHandler struct{}

func (h *SatelliteThermalIngestHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"fire_perimeter_hectares":8420,"active_hotspots":156,"fire_radiative_power_mw":1240.5,"satellite_pass":"NOAA-20"}`),
	}, nil
}

// DEMFuelLayerPrepHandler overlays Anderson 13 fuel model categories and slope rasters.
type DEMFuelLayerPrepHandler struct{}

func (h *DEMFuelLayerPrepHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"fuel_model":"SH9_HIGH_LOAD_SHRUB","mean_slope_pct":32.4,"aspect_deg":215,"fuel_moisture_pct":4.2}`),
	}, nil
}

// RothermelROSCalcHandler computes forward rate of fire spread under ambient wind gusts.
type RothermelROSCalcHandler struct{}

func (h *RothermelROSCalcHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"head_ros_meters_min":18.4,"flame_length_meters":4.2,"reaction_intensity_kw_m2":8900,"fire_spread_direction_deg":48}`),
	}, nil
}

// EmberSpottingSimHandler simulates aerodynamic lofting and ignition of downwind spotting embers.
type EmberSpottingSimHandler struct{}

func (h *EmberSpottingSimHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"spot_fire_probability_pct":78.5,"max_spot_distance_km":2.4,"simulated_trajectories":10000,"breached_containment":true}`),
	}, nil
}

// IPAWSEvacuationBroadcastHandler signs and emits Common Alerting Protocol (CAP) XML messages to civil defense.
type IPAWSEvacuationBroadcastHandler struct{}

func (h *IPAWSEvacuationBroadcastHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"cap_alert_id":"IPAWS-OR-2026-FIRE-04","alert_status":"ORDER_EVACUATION_LEVEL_3","zones_affected":["ZONE_4A","ZONE_4B"],"recipient_count":14200,"status":"EMERGENCY_BROADCAST_SENT"}`),
	}, nil
}
