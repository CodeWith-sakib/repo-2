package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestOffshoreWindTurbineFarmPipeline_Validate(t *testing.T) {
	wf := OffshoreWindTurbineFarmPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestOffshoreWindTurbineFarmPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-wind-01", StepID: "nacelle-lidar-wind-vector-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"lidar_wind_vector_ingest", &LiDARWindVectorHandler{}},
		{"wake_deficit_sim", &WakeDeficitSimHandler{}},
		{"active_yaw_steering_opt", &ActiveYawSteeringHandler{}},
		{"gearbox_acoustic_check", &GearboxAcousticHandler{}},
		{"hvdc_inverter_dispatch", &HVDCInverterDispatchHandler{}},
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
