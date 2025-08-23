package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestSupplyChainTrackingPipeline(t *testing.T) {
	pipeline := SupplyChainTrackingPipeline()
	if err := pipeline.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}

	ctx := context.Background()
	dummyStepCtx := worker.StepContext{
		RunID:    "run_sc_01",
		StepID:   "ingest-manifest",
		WorkerID: "worker-01",
	}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"manifest", &ManifestIngestHandler{}},
		{"customs", &CustomsVerifierHandler{}},
		{"telematics", &TelematicsAuditHandler{}},
		{"route", &RoutePlannerHandler{}},
		{"dock", &DockSchedulerHandler{}},
		{"ebol", &EBOLSignerHandler{}},
	}

	for _, h := range handlers {
		res, err := h.handler.Execute(ctx, dummyStepCtx)
		if err != nil {
			t.Fatalf("handler %s returned error: %v", h.name, err)
		}
		if len(res.Output) == 0 {
			t.Errorf("handler %s returned empty output", h.name)
		}
	}
}
