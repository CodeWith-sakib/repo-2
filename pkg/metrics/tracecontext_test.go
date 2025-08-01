package metrics

import (
	"testing"
)

func TestParseTraceparent(t *testing.T) {
	header := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	tp, err := ParseTraceparent(header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tp.TraceID != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Errorf("unexpected trace id: %s", tp.TraceID)
	}
	if tp.ParentID != "00f067aa0ba902b7" {
		t.Errorf("unexpected parent id: %s", tp.ParentID)
	}
	if tp.String() != header {
		t.Errorf("expected string serialization roundtrip %s, got %s", header, tp.String())
	}
}
