package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestSmartGridEnergyPipeline_Validate(t *testing.T) {
	pipeline := SmartGridEnergyPipeline()
	if err := pipeline.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
}

func TestSmartGridEnergyPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-grid-01", StepID: "smart-meter-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"meter", &SmartMeterIngestHandler{}},
		{"forecast", &RenewableForecastHandler{}},
		{"curtail", &LoadCurtailmentOptimizerHandler{}},
		{"bess", &BESSBatteryDispatchHandler{}},
		{"spot", &SpotMarketBidderHandler{}},
		{"freq", &GridFrequencyStabilizerHandler{}},
		{"settle", &UtilitySettlementLedgerHandler{}},
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
