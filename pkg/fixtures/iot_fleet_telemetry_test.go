package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestIoTFleetTelemetryPipeline_Validate(t *testing.T) {
	pipeline := IoTFleetTelemetryPipeline()
	if err := pipeline.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
}

func TestIoTFleetTelemetryPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-iot-01", StepID: "mqtt-sensor-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"ingest", &MQTTSensorIngestHandler{}},
		{"geofence", &GeofenceValidatorHandler{}},
		{"thermal", &BatteryThermalMonitorHandler{}},
		{"speed", &SpeedTelematicsScorerHandler{}},
		{"ota", &FirmwareOTACheckerHandler{}},
		{"trip", &TripSummaryPersisterHandler{}},
		{"alert", &FleetAlertNotifierHandler{}},
	}

	for _, h := range handlers {
		res, err := h.handler.Execute(ctx, sctx)
		if err != nil {
			t.Fatalf("handler %s failed: %v", h.name, err)
		}
		if len(res.Output) == 0 {
			t.Errorf("handler %s returned empty output", h.name)
		}
	}
}
