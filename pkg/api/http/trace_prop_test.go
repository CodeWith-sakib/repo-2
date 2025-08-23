package http

import (
	"net/http/httptest"
	"testing"
)

func TestTracePropagator_ExtractAndInject(t *testing.T) {
	prop := NewTracePropagator()

	req := httptest.NewRequest("GET", "/api/v1/workflows", nil)
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	req.Header.Set("baggage", "userId=alice,role=admin")

	extracted, err := prop.Extract(req)
	if err != nil {
		t.Fatalf("failed to extract traceparent: %v", err)
	}
	if extracted == nil {
		t.Fatalf("expected trace extracted, got nil")
	}

	if extracted.TraceID != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Errorf("unexpected trace ID: %s", extracted.TraceID)
	}
	if extracted.BaggageItems["userId"] != "alice" || extracted.BaggageItems["role"] != "admin" {
		t.Errorf("unexpected baggage items: %+v", extracted.BaggageItems)
	}

	outReq := httptest.NewRequest("POST", "/downstream", nil)
	prop.Inject(outReq, extracted)

	if outReq.Header.Get("traceparent") != "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01" {
		t.Errorf("injected traceparent mismatch: %s", outReq.Header.Get("traceparent"))
	}
}
