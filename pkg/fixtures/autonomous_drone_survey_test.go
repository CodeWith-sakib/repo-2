package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestAutonomousDroneSurveyPipeline_Validate(t *testing.T) {
	pipeline := AutonomousDroneSurveyPipeline()
	if err := pipeline.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
}

func TestAutonomousDroneSurveyPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-drone-01", StepID: "telemetry-lidar-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"ingest", &DroneIngestHandler{}},
		{"airspace", &AirspaceCheckHandler{}},
		{"safety", &SafetyMarginsHandler{}},
		{"sfm", &SfMReconstructHandler{}},
		{"thermal", &ThermalDetectHandler{}},
		{"mesh", &MeshSurfaceHandler{}},
		{"report", &DroneReportGenerateHandler{}},
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
