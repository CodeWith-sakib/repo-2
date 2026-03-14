package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestGeospatialWildfireSpreadPipeline_Validate(t *testing.T) {
	wf := GeospatialWildfireSpreadPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestGeospatialWildfireSpreadPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-fire-01", StepID: "modis-viirs-satellite-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"satellite_thermal_ingest", &SatelliteThermalIngestHandler{}},
		{"dem_fuel_layer_prep", &DEMFuelLayerPrepHandler{}},
		{"rothermel_ros_calc", &RothermelROSCalcHandler{}},
		{"ember_spotting_sim", &EmberSpottingSimHandler{}},
		{"ipaws_evacuation_broadcast", &IPAWSEvacuationBroadcastHandler{}},
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
