package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestAutonomousDroneFleetMissionPipeline_Validate(t *testing.T) {
	wf := AutonomousDroneFleetMissionPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestAutonomousDroneFleetMissionPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-sar-uav-01", StepID: "geofence-and-weather-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"geofence_weather_ingest", &GeofenceWeatherHandler{}},
		{"dubins_path_planner", &DubinsPathPlannerHandler{}},
		{"swarm_boids_collision", &SwarmBoidsCollisionHandler{}},
		{"flir_thermal_detector", &FLIRThermalDetectorHandler{}},
		{"incident_command_relay", &IncidentCommandRelayHandler{}},
	}

	for _, h := range handlers {
		res, err := h.handler.Execute(ctx, sctx)
		if err != nil {
			t.Fatalf("handler %s failed: %v", h.name, err)
		}
		if len(res.Output) == 0 {
			t.Errorf("expected non-empty output for handler %s", h.name)
		}
	}
}
