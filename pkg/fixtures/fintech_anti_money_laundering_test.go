package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestAMLSanctionsPipeline_Validate(t *testing.T) {
	pipeline := AMLSanctionsPipeline()
	if err := pipeline.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
}

func TestAMLSanctionsPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-aml-01", StepID: "transaction-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"ingest", &AMLTransactionIngestHandler{}},
		{"ofac", &OFACSanctionsLookupHandler{}},
		{"pep", &PEPScreenerHandler{}},
		{"structuring", &StructuringDetectorHandler{}},
		{"risk_score", &RiskScoreAggregatorHandler{}},
		{"sar", &SARGeneratorHandler{}},
		{"case", &ComplianceCaseCreationHandler{}},
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
