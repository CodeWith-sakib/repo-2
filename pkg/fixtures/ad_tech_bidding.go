package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// AdTechBiddingPipeline builds a real-time ad exchange bidding DAG.
func AdTechBiddingPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_ad_bidding"),
		TenantID:    "ad-exchange",
		Name:        "Real-Time Bidding (RTB) Auction Pipeline",
		Version:     1,
		Description: "Ultra-low latency OpenRTB bid request parsing, user profile matching, budget pacing, second-price auction clearing, and win notice dispatch",
		Timeout:     30 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "bid-request-parser",
				TaskType: "rtb_parser",
			},
			{
				ID:        "user-profile-enricher",
				TaskType:  "profile_enrich",
				DependsOn: []string{"bid-request-parser"},
			},
			{
				ID:        "budget-pacing-checker",
				TaskType:  "budget_pacing",
				DependsOn: []string{"user-profile-enricher"},
			},
			{
				ID:        "contextual-targeting-matcher",
				TaskType:  "targeting_match",
				DependsOn: []string{"budget-pacing-checker"},
			},
			{
				ID:        "auction-clearing-engine",
				TaskType:  "auction_clear",
				DependsOn: []string{"contextual-targeting-matcher"},
			},
			{
				ID:        "creative-asset-assembler",
				TaskType:  "creative_assembler",
				DependsOn: []string{"auction-clearing-engine"},
			},
			{
				ID:        "win-notice-dispatcher",
				TaskType:  "win_notice",
				DependsOn: []string{"creative-asset-assembler"},
			},
		},
	}
}

// RTBParserHandler parses OpenRTB 2.5 bid request protocols.
type RTBParserHandler struct{}

func (h *RTBParserHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"bid_request_id":"br-90210","imp_count":1,"slot_size":"300x250","floor_cpm_usd":1.50,"device":"mobile_ios"}`),
	}, nil
}

// UserProfileEnricherHandler looks up DMP audience segments and demographic cohorts.
type UserProfileEnricherHandler struct{}

func (h *UserProfileEnricherHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"user_token":"anon-usr-77","segments":["tech_enthusiast","auto_intender"],"frequency_24h":4,"geo_country":"US"}`),
	}, nil
}

// BudgetPacingCheckerHandler evaluates campaign daily pacing burn rates using PID controllers.
type BudgetPacingCheckerHandler struct{}

func (h *BudgetPacingCheckerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"eligible_campaigns":14,"budget_exhausted_campaigns":2,"pacing_multiplier":0.92}`),
	}, nil
}

// TargetingMatcherHandler filters candidate creatives matching slot dimensions and keywords.
type TargetingMatcherHandler struct{}

func (h *TargetingMatcherHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"matched_bids":[{"dsp":"DSP-A","bid_cpm":3.20},{"dsp":"DSP-B","bid_cpm":2.85},{"dsp":"DSP-C","bid_cpm":1.90}]}`),
	}, nil
}

// AuctionClearingEngineHandler runs Vickrey second-price auction mechanics.
type AuctionClearingEngineHandler struct{}

func (h *AuctionClearingEngineHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"winner_dsp":"DSP-A","highest_bid_cpm":3.20,"clearing_price_cpm":2.86,"auction_type":"second_price"}`),
	}, nil
}

// CreativeAssemblerHandler packages HTML5 creative payload, tracking pixels, and click tags.
type CreativeAssemblerHandler struct{}

func (h *CreativeAssemblerHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"adm":"<script src='https://cdn.ad.net/banner.js'></script>","click_url":"https://tracker.net/click?id=44","markup_size_bytes":1420}`),
	}, nil
}

// WinNoticeDispatcherHandler asynchronously posts OpenRTB nurl win beacons to DSPs.
type WinNoticeDispatcherHandler struct{}

func (h *WinNoticeDispatcherHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	noticeID := fmt.Sprintf("win-%d", time.Now().UnixNano())
	payload := fmt.Sprintf(`{"notice_id":"%s","dsp_endpoint":"https://dsp-a.net/win","settled_usd":0.00286,"status":"dispatched"}`, noticeID)
	return &worker.StepResult{Output: json.RawMessage(payload)}, nil
}
