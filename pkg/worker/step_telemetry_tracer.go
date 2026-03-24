package worker

import (
	"context"
	"sync"
	"time"
)

// StepSpan captures distributed trace timing across worker step execution.
type StepSpan struct {
	SpanID    string        `json:"span_id"`
	TraceID   string        `json:"trace_id"`
	StepID    string        `json:"step_id"`
	StartTime time.Time     `json:"start_time"`
	Duration  time.Duration `json:"duration"`
	Tags      map[string]string `json:"tags"`
	Status    string        `json:"status"`
}

// StepTelemetryTracer manages in-memory spans and timing lifecycle.
type StepTelemetryTracer struct {
	mu    sync.Mutex
	spans []StepSpan
}

// NewStepTelemetryTracer creates a telemetry tracer.
func NewStepTelemetryTracer() *StepTelemetryTracer {
	return &StepTelemetryTracer{
		spans: make([]StepSpan, 0, 100),
	}
}

// StartSpan opens a new timing span.
func (t *StepTelemetryTracer) StartSpan(ctx context.Context, traceID, spanID, stepID string) *StepSpan {
	return &StepSpan{
		SpanID:    spanID,
		TraceID:   traceID,
		StepID:    stepID,
		StartTime: time.Now().UTC(),
		Tags:      make(map[string]string),
		Status:    "ACTIVE",
	}
}

// EndSpan finalizes span duration and captures it in the trace buffer.
func (t *StepTelemetryTracer) EndSpan(span *StepSpan, status string) {
	if span == nil {
		return
	}
	span.Duration = time.Since(span.StartTime)
	span.Status = status

	t.mu.Lock()
	defer t.mu.Unlock()
	t.spans = append(t.spans, *span)
}

// Spans returns copy of captured telemetry spans.
func (t *StepTelemetryTracer) Spans() []StepSpan {
	t.mu.Lock()
	defer t.mu.Unlock()

	res := make([]StepSpan, len(t.spans))
	copy(res, t.spans)
	return res
}
