package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// AutonomousDroneFleetMissionPipeline builds a search-and-rescue UAV swarm mission DAG.
func AutonomousDroneFleetMissionPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_uav_swarm_mission"),
		TenantID:    "search-and-rescue-command",
		Name:        "Autonomous UAV Swarm Search-and-Rescue Flight Mission DAG",
		Version:     1,
		Description: "Ingests geofence boundaries, calculates 3D Dubins path trajectory, solves decentralized collision avoidance (Boids flocking model), executes FLIR thermal anomaly detection, and relays situational coordinates to incident command.",
		Timeout:     180 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "geofence-and-weather-ingest",
				TaskType: "geofence_weather_ingest",
			},
			{
				ID:        "3d-dubins-path-planning",
				TaskType:  "dubins_path_planner",
				DependsOn: []string{"geofence-and-weather-ingest"},
			},
			{
				ID:        "swarm-collision-avoidance",
				TaskType:  "swarm_boids_collision",
				DependsOn: []string{"3d-dubins-path-planning"},
			},
			{
				ID:        "flir-thermal-anomaly-scan",
				TaskType:  "flir_thermal_detector",
				DependsOn: []string{"swarm-collision-avoidance"},
			},
			{
				ID:        "incident-command-relay",
				TaskType:  "incident_command_relay",
				DependsOn: []string{"flir-thermal-anomaly-scan"},
			},
		},
	}
}

// GeofenceWeatherHandler validates no-fly zones and wind gust thresholds.
type GeofenceWeatherHandler struct{}

func (h *GeofenceWeatherHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"search_area_km2":45.2,"wind_speed_knots":12.4,"no_fly_zones_cleared":true,"visibility_meters":10000}`),
	}, nil
}

// DubinsPathPlannerHandler generates curvature-constrained Dubins flight trajectories.
type DubinsPathPlannerHandler struct{}

func (h *DubinsPathPlannerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"waypoints_count":148,"total_flight_distance_km":38.6,"cruising_speed_mps":18.0,"min_turning_radius_m":15.0}`),
	}, nil
}

// SwarmBoidsCollisionHandler applies Reynolds boids flocking rules (separation, alignment, cohesion).
type SwarmBoidsCollisionHandler struct{}

func (h *SwarmBoidsCollisionHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"active_uavs":8,"min_inter_uav_distance_m":24.5,"collision_risk":"LOW","formation":"EXTENDED_SWEEP"}`),
	}, nil
}

// FLIRThermalDetectorHandler runs edge deep learning model on radiometric infrared video feed.
type FLIRThermalDetectorHandler struct{}

func (h *FLIRThermalDetectorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"anomalies_detected":1,"lat":47.6062,"lon":-122.3321,"thermal_signature_c":37.1,"confidence_pct":96.8}`),
	}, nil
}

// IncidentCommandRelayHandler broadcasts encrypted tactical coordinates to ground emergency teams.
type IncidentCommandRelayHandler struct{}

func (h *IncidentCommandRelayHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"dispatch_ticket":"SAR-2026-088","relay_protocol":"ATAK_COTK","status":"CONFIRMED_BY_GROUND_LEAD"}`),
	}, nil
}
