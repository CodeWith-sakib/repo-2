package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestFraudDetectionPipeline(t *testing.T) {
	pipeline := FraudDetectionPipeline()
	if err := pipeline.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}

	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-fraud-01", StepID: "ingest-transaction"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"ingest", &TxIngestHandler{}},
		{"features", &FeatureExtractorHandler{}},
		{"velocity", &VelocityCheckerHandler{}},
		{"geo", &GeoAnomalyHandler{}},
		{"ml", &MLScorerHandler{}},
		{"rules", &RuleEvaluatorHandler{}},
		{"aggregator", &RiskAggregatorHandler{}},
		{"decision", &DecisionHandler{}},
	}

	for _, h := range handlers {
		res, err := h.handler.Execute(ctx, sctx)
		if err != nil {
			t.Fatalf("handler %s error: %v", h.name, err)
		}
		if len(res.Output) == 0 {
			t.Errorf("handler %s returned empty output", h.name)
		}
	}
}
