package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestCryogenicMagnetQuenchProtectionPipeline_Validate(t *testing.T) {
	wf := CryogenicMagnetQuenchProtectionPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestCryogenicMagnetQuenchProtectionPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-cryo-01", StepID: "bridge-voltage-coil-telemetry-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"coil_voltage_ingest", &CoilVoltageIngestHandler{}},
		{"quench_transient_detect", &QuenchTransientDetectHandler{}},
		{"quench_heater_fire", &QuenchHeaterFireHandler{}},
		{"thyristor_breaker_trip", &ThyristorBreakerTripHandler{}},
		{"energy_dump_dissipate", &EnergyDumpDissipateHandler{}},
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
