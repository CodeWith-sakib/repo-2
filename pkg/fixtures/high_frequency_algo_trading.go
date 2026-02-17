package fixtures

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// HighFrequencyAlgoTradingPipeline builds a quantitative market microstructure execution DAG.
func HighFrequencyAlgoTradingPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_hft_market_making"),
		TenantID:    "hft-quant-desk",
		Name:        "High-Frequency Quantitative Market Making & Order Book Risk Pipeline",
		Version:     1,
		Description: "Parses ITCH 5.0 L3 limit order book ticks, computes volume-synchronized probability of toxicity (VPIN), updates Hawkes process order arrival intensities, solves Avellaneda-Stoikov inventory skew, and transmits FIX 4.4 OUCH execution cancel/replace orders.",
		Timeout:     30 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "l3-itch-orderbook-ingest",
				TaskType: "itch_l3_ingest",
			},
			{
				ID:        "vpin-toxicity-calculation",
				TaskType:  "vpin_toxicity_calc",
				DependsOn: []string{"l3-itch-orderbook-ingest"},
			},
			{
				ID:        "hawkes-intensity-modeling",
				TaskType:  "hawkes_intensity_model",
				DependsOn: []string{"l3-itch-orderbook-ingest"},
			},
			{
				ID:        "avellaneda-stoikov-quote-skew",
				TaskType:  "as_quote_skew_solver",
				DependsOn: []string{"vpin-toxicity-calculation", "hawkes-intensity-modeling"},
			},
			{
				ID:        "ouch-order-cancel-replace-transmit",
				TaskType:  "ouch_order_transmit",
				DependsOn: []string{"avellaneda-stoikov-quote-skew"},
			},
		},
	}
}

// ITCHOrderBookHandler processes nanosecond level-3 market data feeds.
type ITCHOrderBookHandler struct{}

func (h *ITCHOrderBookHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"symbol":"NVDA","best_bid":124.50,"best_ask":124.51,"spread_ticks":1,"l3_depth_levels":20,"timestamp_ns":1717142400000000}`),
	}, nil
}

// VPINToxicityHandler computes flow toxicity metric based on volume bucket imbalances.
type VPINToxicityHandler struct{}

func (h *VPINToxicityHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"vpin_metric":0.28,"toxic_flow_detected":false,"volume_bucket_size":50000,"imbalance_ratio":0.12}`),
	}, nil
}

// HawkesIntensityHandler estimates mutually exciting point process order arrival rates.
type HawkesIntensityHandler struct{}

func (h *HawkesIntensityHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"baseline_intensity_lambda0":42.5,"self_excitation_alpha":0.82,"decay_rate_beta":1.45,"order_burst_probability":0.15}`),
	}, nil
}

// ASQuoteSkewHandler calculates optimal bid/ask reservation price offset under inventory risk.
type ASQuoteSkewHandler struct{}

func (h *ASQuoteSkewHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"current_inventory_shares":-400,"reservation_price":124.508,"optimal_bid":124.50,"optimal_ask":124.52,"skew_applied":"BID_PASSIVE"}`),
	}, nil
}

// OUCHOrderTransmitHandler emits binary low-latency cancel/replace orders.
type OUCHOrderTransmitHandler struct{}

func (h *OUCHOrderTransmitHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"clordid":"ORD-2026-NVDA-991","order_action":"REPLACE_PRICE","firm_id":"HFT-DESK-1","wire_latency_ns":820,"status":"ACCEPTED"}`),
	}, nil
}
