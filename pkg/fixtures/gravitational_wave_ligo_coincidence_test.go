package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestGravitationalWaveLIGOCoincidencePipeline_Validate(t *testing.T) {
	wf := GravitationalWaveLIGOCoincidencePipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestGravitationalWaveLIGOCoincidencePipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-ligo-01", StepID: "hanford-livingston-strain-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"strain_series_ingest", &StrainSeriesIngestHandler{}},
		{"qtransform_whitening", &QTransformWhiteningHandler{}},
		{"matched_filter_bank", &MatchedFilterBankHandler{}},
		{"temporal_coincidence_eval", &TemporalCoincidenceEvalHandler{}},
		{"gcn_skymap_broadcast", &GCNSkymapBroadcastHandler{}},
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
