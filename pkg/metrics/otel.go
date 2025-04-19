package metrics

import (
	"context"
	"encoding/hex"
	"math/rand"
	"sync"
	"time"
)

type SpanKind int

const (
	SpanKindInternal SpanKind = iota
	SpanKindServer
	SpanKindClient
	SpanKindProducer
	SpanKindConsumer
)

type SpanContext struct {
	TraceID string `json:"trace_id"`
	SpanID  string `json:"span_id"`
}

type SpanRecord struct {
	Name       string                 `json:"name"`
	Context    SpanContext            `json:"context"`
	ParentID   string                 `json:"parent_id,omitempty"`
	Kind       SpanKind               `json:"kind"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    time.Time              `json:"end_time"`
	Duration   time.Duration          `json:"duration_ms"`
	Attributes map[string]interface{} `json:"attributes"`
	Status     string                 `json:"status"`
}

type TraceExporter interface {
	ExportSpans(ctx context.Context, spans []*SpanRecord) error
}

type MemorySpanExporter struct {
	mu    sync.RWMutex
	spans []*SpanRecord
}

func NewMemorySpanExporter() *MemorySpanExporter {
	return &MemorySpanExporter{
		spans: make([]*SpanRecord, 0),
	}
}

func (e *MemorySpanExporter) ExportSpans(ctx context.Context, spans []*SpanRecord) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.spans = append(e.spans, spans...)
	return nil
}

func (e *MemorySpanExporter) GetSpans() []*SpanRecord {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]*SpanRecord, len(e.spans))
	copy(out, e.spans)
	return out
}

type Tracer struct {
	serviceName string
	exporter    TraceExporter
}

func NewTracer(serviceName string, exporter TraceExporter) *Tracer {
	return &Tracer{
		serviceName: serviceName,
		exporter:    exporter,
	}
}

func (t *Tracer) StartSpan(name string, parent *SpanContext) (*SpanRecord, func(status string, attrs map[string]interface{})) {
	traceID := ""
	parentID := ""
	if parent != nil {
		traceID = parent.TraceID
		parentID = parent.SpanID
	} else {
		traceID = randomHex(16)
	}
	spanID := randomHex(8)

	rec := &SpanRecord{
		Name: name,
		Context: SpanContext{
			TraceID: traceID,
			SpanID:  spanID,
		},
		ParentID:   parentID,
		StartTime:  time.Now().UTC(),
		Attributes: make(map[string]interface{}),
	}
	rec.Attributes["service.name"] = t.serviceName

	endFn := func(status string, attrs map[string]interface{}) {
		rec.EndTime = time.Now().UTC()
		rec.Duration = rec.EndTime.Sub(rec.StartTime)
		rec.Status = status
		for k, v := range attrs {
			rec.Attributes[k] = v
		}
		if t.exporter != nil {
			_ = t.exporter.ExportSpans(context.Background(), []*SpanRecord{rec})
		}
	}

	return rec, endFn
}

func randomHex(bytesLen int) string {
	b := make([]byte, bytesLen)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
