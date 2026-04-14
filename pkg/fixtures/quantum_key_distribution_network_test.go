package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestQuantumKeyDistributionPipeline_Validate(t *testing.T) {
	wf := QuantumKeyDistributionPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestQuantumKeyDistributionPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-qkd-01", StepID: "optical-beacon-telescope-tracking"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"telescope_beacon_track", &TelescopeBeaconTrackHandler{}},
		{"bb84_photon_detect", &BB84PhotonDetectHandler{}},
		{"qber_sifting_check", &QBERSiftingCheckHandler{}},
		{"cascade_error_correction", &CascadeErrorCorrectionHandler{}},
		{"toeplitz_privacy_amp", &ToeplitzPrivacyAmpHandler{}},
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
