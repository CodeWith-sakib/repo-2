package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestEnergyGridLoadDispatchPipeline_Validate(t *testing.T) {
	wf := EnergyGridLoadDispatchPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestEnergyGridLoadDispatchPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-grid-01", StepID: "pmu-phasor-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"pmu_ingest", &PMUPhasorHandler{}},
		{"renewable_forecast", &GridRenewableForecastHandler{}},
		{"acopf_solver", &ACOPFHandler{}},
		{"n1_security_analysis", &N1SecurityAnalysisHandler{}},
		{"bess_dispatch", &BESSDispatchHandler{}},
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
