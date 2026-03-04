package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestSmartCityTrafficOptimizationPipeline_Validate(t *testing.T) {
	wf := SmartCityTrafficOptimizationPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestSmartCityTrafficOptimizationPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-traffic-01", StepID: "loop-and-cctv-telemetry-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"traffic_telemetry_ingest", &TrafficTelemetryHandler{}},
		{"queue_length_estimator", &QueueLengthEstimatorHandler{}},
		{"webster_cycle_optimizer", &WebsterCycleOptimizerHandler{}},
		{"green_wave_corridor", &GreenWaveCorridorHandler{}},
		{"vms_signage_broadcast", &VMSSignageBroadcastHandler{}},
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
