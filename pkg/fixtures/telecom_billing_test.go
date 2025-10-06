package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestTelecomBillingPipeline_Validate(t *testing.T) {
	pipeline := TelecomBillingPipeline()
	if err := pipeline.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
}

func TestTelecomBillingPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-telco-01", StepID: "cdr-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"ingest", &CDRIngestHandler{}},
		{"dedup", &CDRDedupHandler{}},
		{"rating", &TariffRatingHandler{}},
		{"discount", &DiscountEngineHandler{}},
		{"tax", &TaxCalculatorHandler{}},
		{"invoice", &InvoiceGeneratorHandler{}},
		{"gl", &GLPostingHandler{}},
	}

	for _, h := range handlers {
		res, err := h.handler.Execute(ctx, sctx)
		if err != nil {
			t.Fatalf("handler %s failed: %v", h.name, err)
		}
		if len(res.Output) == 0 {
			t.Errorf("handler %s produced empty output", h.name)
		}
	}
}
