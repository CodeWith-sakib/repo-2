package events

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestCloudEventConversionRoundtrip(t *testing.T) {
	evt := &core.Event{
		ID:        core.NewID("evt-1"),
		RunID:     core.NewID("run-1"),
		TenantID:  "tenant-alpha",
		Type:      core.EventRunCompleted,
		Timestamp: time.Now().UTC().Truncate(time.Millisecond),
		Payload:   json.RawMessage(`{"duration_ms":1250}`),
	}

	ce, err := ToCloudEvent(evt, "/kestrel/engine/cluster-1")
	if err != nil {
		t.Fatalf("ToCloudEvent failed: %v", err)
	}

	if ce.SpecVersion != "1.0" {
		t.Errorf("expected specversion 1.0, got %s", ce.SpecVersion)
	}

	back, err := FromCloudEvent(ce)
	if err != nil {
		t.Fatalf("FromCloudEvent failed: %v", err)
	}

	if back.ID != evt.ID || back.RunID != evt.RunID || back.TenantID != evt.TenantID {
		t.Errorf("event mismatch after roundtrip: got %+v, want %+v", back, evt)
	}
}
