package worker

import (
	"context"
	"testing"
	"time"
)

func TestStepTelemetryTracer(t *testing.T) {
	tracer := NewStepTelemetryTracer()

	span := tracer.StartSpan(context.Background(), "trc-01", "spn-01", "step-extract")
	span.Tags["worker_pool"] = "gpu-cluster"

	time.Sleep(20 * time.Millisecond)
	tracer.EndSpan(span, "OK")

	spans := tracer.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span recorded, got %d", len(spans))
	}

	if spans[0].Duration < 15*time.Millisecond {
		t.Errorf("duration too short: %v", spans[0].Duration)
	}

	if spans[0].Tags["worker_pool"] != "gpu-cluster" {
		t.Errorf("tag mismatch: %s", spans[0].Tags["worker_pool"])
	}
}
