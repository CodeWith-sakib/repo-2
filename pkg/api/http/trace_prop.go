package http

import (
	"fmt"
	"net/http"
	"strings"
)

// DistributedTraceHeader represents extracted W3C traceparent and baggage information from incoming HTTP calls.
type DistributedTraceHeader struct {
	TraceID      string
	SpanID       string
	TraceFlags   string
	BaggageItems map[string]string
}

// TracePropagator parses and injects W3C TraceContext headers.
type TracePropagator struct{}

// NewTracePropagator creates a new HTTP trace propagator.
func NewTracePropagator() *TracePropagator {
	return &TracePropagator{}
}

// Extract extracts trace information from incoming HTTP headers.
func (p *TracePropagator) Extract(r *http.Request) (*DistributedTraceHeader, error) {
	raw := r.Header.Get("traceparent")
	if raw == "" {
		return nil, nil // trace not present
	}

	parts := strings.Split(raw, "-")
	if len(parts) != 4 || parts[0] != "00" || len(parts[1]) != 32 || len(parts[2]) != 16 || len(parts[3]) != 2 {
		return nil, fmt.Errorf("malformed traceparent header: %s", raw)
	}

	header := &DistributedTraceHeader{
		TraceID:      parts[1],
		SpanID:       parts[2],
		TraceFlags:   parts[3],
		BaggageItems: make(map[string]string),
	}

	baggageRaw := r.Header.Get("baggage")
	if baggageRaw != "" {
		pairs := strings.Split(baggageRaw, ",")
		for _, pair := range pairs {
			kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
			if len(kv) == 2 {
				header.BaggageItems[kv[0]] = kv[1]
			}
		}
	}

	return header, nil
}

// Inject writes distributed trace headers to an outbound HTTP request.
func (p *TracePropagator) Inject(r *http.Request, trace *DistributedTraceHeader) {
	if trace == nil {
		return
	}
	r.Header.Set("traceparent", fmt.Sprintf("00-%s-%s-%s", trace.TraceID, trace.SpanID, trace.TraceFlags))

	if len(trace.BaggageItems) > 0 {
		var parts []string
		for k, v := range trace.BaggageItems {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
		r.Header.Set("baggage", strings.Join(parts, ", "))
	}
}
