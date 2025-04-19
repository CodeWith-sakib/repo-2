package metrics

import (
	"testing"
)

func TestTracerSpanLifecycle(t *testing.T) {
	exp := NewMemorySpanExporter()
	tracer := NewTracer("kestrel-test", exp)

	parentSpan, endParent := tracer.StartSpan("WorkflowExecution", nil)
	_, endChild := tracer.StartSpan("StepExecution", &parentSpan.Context)

	endChild("OK", map[string]interface{}{"step.id": "step-1"})
	endParent("OK", map[string]interface{}{"workflow.id": "wf-1"})

	spans := exp.GetSpans()
	if len(spans) != 2 {
		t.Fatalf("expected 2 spans exported, got %d", len(spans))
	}

	if spans[1].ParentID != spans[0].Context.SpanID {
		// child was ended first, so spans[0] is child, spans[1] is parent
		if spans[0].ParentID != spans[1].Context.SpanID {
			t.Errorf("expected parent-child relationship between spans")
		}
	}
}
