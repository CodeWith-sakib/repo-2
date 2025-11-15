package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestSupplyChainLogisticsPipeline_Validate(t *testing.T) {
	pipeline := SupplyChainLogisticsPipeline()
	if err := pipeline.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
}

func TestSupplyChainLogisticsPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-freight-01", StepID: "container-telemetry-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"telemetry", &ContainerTelemetryIngestHandler{}},
		{"customs", &CustomsClearanceValidatorHandler{}},
		{"cold_chain", &ColdChainAuditorHandler{}},
		{"eta", &RouteETAPredictorHandler{}},
		{"demurrage", &DemurrageCalculatorHandler{}},
		{"bol", &BillOfLadingGeneratorHandler{}},
		{"drayage", &DrayageDispatchNotifierHandler{}},
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
