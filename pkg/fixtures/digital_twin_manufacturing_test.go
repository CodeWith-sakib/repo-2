package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestDigitalTwinManufacturingPipeline_Validate(t *testing.T) {
	wf := DigitalTwinManufacturingPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestDigitalTwinManufacturingPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-cnc-01", StepID: "opc-ua-telemetry-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"opcua_telemetry_ingest", &OPCUATelemetryHandler{}},
		{"vibration_fft", &VibrationFFTHandler{}},
		{"thermal_expansion_model", &ThermalExpansionHandler{}},
		{"rul_estimation", &RULEstimationHandler{}},
		{"erp_work_order_gen", &ERPWorkOrderGenHandler{}},
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
