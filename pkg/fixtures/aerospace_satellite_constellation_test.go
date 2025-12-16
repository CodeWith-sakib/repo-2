package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestAerospaceSatelliteConstellationPipeline_Validate(t *testing.T) {
	wf := AerospaceSatelliteConstellationPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 7 {
		t.Fatalf("expected 7 steps in aerospace constellation pipeline, got %d", len(wf.Steps))
	}
}

func TestAerospaceSatelliteConstellationPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-sat-4082", StepID: "ground-telemetry-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"ground-telemetry-ingest", &GroundTelemetryIngestHandler{} },
		{"spacecraft-health-decom", &HealthDecomHandler{} },
		{"orbit-sgp4-propagation", &SGP4PropagateHandler{} },
		{"space-debris-conjunction-cdm", &ConjunctionCDMHandler{} },
		{"reaction-wheel-desat", &AttitudeDesatHandler{} },
		{"payload-optical-tasking", &PayloadTaskingHandler{} },
		{"pass-uplink-command-package", &UplinkPackagingHandler{} },
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
