package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestHighFrequencyAlgoTradingPipeline_Validate(t *testing.T) {
	wf := HighFrequencyAlgoTradingPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestHighFrequencyAlgoTradingPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-hft-01", StepID: "l3-itch-orderbook-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"itch_l3_ingest", &ITCHOrderBookHandler{}},
		{"vpin_toxicity_calc", &VPINToxicityHandler{}},
		{"hawkes_intensity_model", &HawkesIntensityHandler{}},
		{"as_quote_skew_solver", &ASQuoteSkewHandler{}},
		{"ouch_order_transmit", &OUCHOrderTransmitHandler{}},
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
