package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// IoTFleetTelemetryPipeline builds an IoT edge fleet sensor aggregation DAG.
func IoTFleetTelemetryPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_iot_telemetry"),
		TenantID:    "auto-fleet-os",
		Name:        "Connected Vehicle Telematics & Fleet Monitoring",
		Version:     1,
		Description: "Real-time MQTT CAN-bus ingest, polygon geofencing, battery thermal anomaly check, driver safety scoring, OTA firmware audit, and fleet alert dispatch",
		Timeout:     45 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "mqtt-sensor-ingest",
				TaskType: "mqtt_ingest",
			},
			{
				ID:        "geofence-validator",
				TaskType:  "geofence",
				DependsOn: []string{"mqtt-sensor-ingest"},
			},
			{
				ID:        "battery-thermal-monitor",
				TaskType:  "thermal_check",
				DependsOn: []string{"geofence-validator"},
			},
			{
				ID:        "speed-telematics-scorer",
				TaskType:  "telematics_score",
				DependsOn: []string{"battery-thermal-monitor"},
			},
			{
				ID:        "firmware-ota-checker",
				TaskType:  "ota_audit",
				DependsOn: []string{"speed-telematics-scorer"},
			},
			{
				ID:        "trip-summary-persister",
				TaskType:  "trip_persist",
				DependsOn: []string{"firmware-ota-checker"},
			},
			{
				ID:        "fleet-alert-notifier",
				TaskType:  "alert_notify",
				DependsOn: []string{"trip-summary-persister"},
			},
		},
	}
}

// MQTTSensorIngestHandler decodes binary Protobuf / JSON vehicle telemetry frames.
type MQTTSensorIngestHandler struct{}

func (h *MQTTSensorIngestHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"vin":"1HGCR2F83HA001234","packets_ingested":150,"protocol":"mqtt_v5","topic":"telemetry/vehicles/vin-1234"}`),
	}, nil
}

// GeofenceValidatorHandler performs point-in-polygon checks against service territories.
type GeofenceValidatorHandler struct{}

func (h *GeofenceValidatorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"lat":37.7749,"lon":-122.4194,"geofence_id":"zone-sf-metro","within_boundary":true}`),
	}, nil
}

// BatteryThermalMonitorHandler detects high cell delta-V and temperature runaways.
type BatteryThermalMonitorHandler struct{}

func (h *BatteryThermalMonitorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"pack_temp_celsius":34.2,"max_cell_temp_celsius":36.1,"state_of_charge_pct":78.4,"thermal_status":"nominal"}`),
	}, nil
}

// SpeedTelematicsScorerHandler computes harsh acceleration, cornering g-force, and speed compliance.
type SpeedTelematicsScorerHandler struct{}

func (h *SpeedTelematicsScorerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"harsh_braking_events":0,"speed_limit_mph":65,"avg_speed_mph":58.4,"safety_score":94.5}`),
	}, nil
}

// FirmwareOTACheckerHandler determines if vehicle ECU is eligible for software update.
type FirmwareOTACheckerHandler struct{}

func (h *FirmwareOTACheckerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"current_firmware":"v4.1.2","target_firmware":"v4.2.0","ota_pending":true,"update_eligible":true}`),
	}, nil
}

// TripSummaryPersisterHandler writes compressed trip time-series logs into Parquet lakehouse.
type TripSummaryPersisterHandler struct{}

func (h *TripSummaryPersisterHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"trip_id":"trip-883492","distance_km":42.5,"duration_minutes":38,"rows_persisted":2280}`),
	}, nil
}

// FleetAlertNotifierHandler sends push notification or PagerDuty incident if abnormal signals detected.
type FleetAlertNotifierHandler struct{}

func (h *FleetAlertNotifierHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	alertID := fmt.Sprintf("alt-%d", time.Now().UnixNano())
	payload := fmt.Sprintf(`{"alert_id":"%s","severity":"INFO","message":"Trip completed within safety parameters","notified":false}`, alertID)
	return &worker.StepResult{Output: json.RawMessage(payload)}, nil
}
