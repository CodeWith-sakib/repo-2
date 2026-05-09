package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestDeepSeaOceanographicSubmersiblePipeline_Validate(t *testing.T) {
	wf := DeepSeaOceanographicSubmersiblePipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestDeepSeaOceanographicSubmersiblePipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-auv-01", StepID: "multibeam-sonar-bathymetry-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"multibeam_sonar_ingest", &MultibeamSonarIngestHandler{}},
		{"dvl_inertial_navigation", &DVLInertialNavigationHandler{}},
		{"plume_gradient_trace", &PlumeGradientTraceHandler{}},
		{"photogrammetry_benthic_map", &PhotogrammetryBenthicMapHandler{}},
		{"usbl_ascent_vector_calc", &USBLAscentVectorHandler{}},
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
