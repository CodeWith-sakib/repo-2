package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestTelecom5GNetworkSlicingPipeline_Validate(t *testing.T) {
	wf := Telecom5GNetworkSlicingPipeline()
	if err := wf.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
	if len(wf.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(wf.Steps))
	}
}

func TestTelecom5GNetworkSlicingPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-5g-slice-01", StepID: "ran-cudu-telemetry-ingest"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"cudu_telemetry_ingest", &CUDUTelemetryHandler{}},
		{"slice_admission_control", &SliceAdmissionHandler{}},
		{"upf_dpdk_configure", &UPFDPDKHandler{}},
		{"sdn_routing_program", &SDNRoutingHandler{}},
		{"sla_latency_assurance", &SLALatencyAssuranceHandler{}},
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
