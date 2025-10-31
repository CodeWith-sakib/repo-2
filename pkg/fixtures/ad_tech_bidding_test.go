package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestAdTechBiddingPipeline_Validate(t *testing.T) {
	pipeline := AdTechBiddingPipeline()
	if err := pipeline.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
}

func TestAdTechBiddingPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-rtb-01", StepID: "bid-request-parser"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"parser", &RTBParserHandler{}},
		{"enrich", &UserProfileEnricherHandler{}},
		{"pacing", &BudgetPacingCheckerHandler{}},
		{"match", &TargetingMatcherHandler{}},
		{"auction", &AuctionClearingEngineHandler{}},
		{"creative", &CreativeAssemblerHandler{}},
		{"win_notice", &WinNoticeDispatcherHandler{}},
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
