package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// AerospaceSatelliteConstellationPipeline generates a multi-stage LEO satellite telemetry & orbit ephemeris DAG.
func AerospaceSatelliteConstellationPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_leo_satellite_ephemeris"),
		TenantID:    "aerospace-orbital-ops",
		Name:        "LEO Satellite Orbit Determination & Ground Station Telemetry Pass DAG",
		Version:     1,
		Description: "Ingests S-band ground station telemetry frames, performs orbit ephemeris propagation, calculates collision avoidance Conjunction Data Messages (CDMs), and schedules next ground contact pass.",
		Timeout:     300 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "ground-telemetry-ingest",
				TaskType: "telemetry_ingest",
			},
			{
				ID:        "spacecraft-health-decom",
				TaskType:  "health_decom",
				DependsOn: []string{"ground-telemetry-ingest"},
			},
			{
				ID:        "orbit-sgp4-propagation",
				TaskType:  "sgp4_propagate",
				DependsOn: []string{"ground-telemetry-ingest"},
			},
			{
				ID:        "space-debris-conjunction-cdm",
				TaskType:  "conjunction_cdm",
				DependsOn: []string{"orbit-sgp4-propagation"},
			},
			{
				ID:        "reaction-wheel-desat",
				TaskType:  "attitude_desat",
				DependsOn: []string{"spacecraft-health-decom"},
			},
			{
				ID:        "payload-optical-tasking",
				TaskType:  "payload_tasking",
				DependsOn: []string{"orbit-sgp4-propagation"},
			},
			{
				ID:        "pass-uplink-command-package",
				TaskType:  "uplink_packaging",
				DependsOn: []string{"space-debris-conjunction-cdm", "reaction-wheel-desat", "payload-optical-tasking"},
			},
		},
	}
}

// GroundTelemetryIngestHandler ingests raw downlink telemetry packets from Svalbard ground station.
type GroundTelemetryIngestHandler struct{}

func (h *GroundTelemetryIngestHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"spacecraft_id":"SAT-LEO-4082","pass_station":"Svalbard-SG3","frames_total":148500,"crc_errors":0,"status":"TELEMETRY_INGESTED"}`),
	}, nil
}

// HealthDecomHandler verifies power, thermal, and reaction wheel state vectors.
type HealthDecomHandler struct{}

func (h *HealthDecomHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"battery_soc_pct":94.2,"solar_panel_v":28.4,"bus_temp_celsius":18.6,"anomaly_detected":false}`),
	}, nil
}

// SGP4PropagateHandler propagates orbital two-line elements (TLE) across the planning horizon.
type SGP4PropagateHandler struct{}

func (h *SGP4PropagateHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"semi_major_axis_km":6878.14,"eccentricity":0.0012,"inclination_deg":97.4,"altitude_km":508.3}`),
	}, nil
}

// ConjunctionCDMHandler checks collision probability against 18th Space Defense Squadron catalog.
type ConjunctionCDMHandler struct{}

func (h *ConjunctionCDMHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"cdm_events_count":1,"closest_approach":"2026-06-02T14:22:10Z","miss_distance_m":4250.0,"maneuver_needed":false}`),
	}, nil
}

// AttitudeDesatHandler schedules magnetic torquer firing during eclipse to desaturate momentum wheels.
type AttitudeDesatHandler struct{}

func (h *AttitudeDesatHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"torquer_firing_seconds":14.5,"momentum_dump_status":"COMPLETE"}`),
	}, nil
}

// PayloadTaskingHandler calculates optical sensor field-of-view ground tracks.
type PayloadTaskingHandler struct{}

func (h *PayloadTaskingHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"target_coordinates":"45.109,-122.341","sun_elevation_deg":58.2,"duty_cycle_sec":45}`),
	}, nil
}

// UplinkPackagingHandler creates and cryptographically signs telecommands for uplink pass.
type UplinkPackagingHandler struct{}

func (h *UplinkPackagingHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"package_id":"CMD-PKG-20260530-01","commands_count":28,"crypto_sig":"ed25519-valid","status":"READY_FOR_UPLINK"}`),
	}, nil
}
