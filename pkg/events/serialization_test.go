package events

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestEventRoundTripSerialization(t *testing.T) {
	// Must test every domain event type for round-trip fidelity
	allEventTypes := []core.EventType{
		core.EventWorkflowCreated,
		core.EventWorkflowUpdated,
		core.EventWorkflowDeleted,
		core.EventRunStarted,
		core.EventRunCompleted,
		core.EventRunFailed,
		core.EventRunCancelled,
		core.EventRunSuspended,
		core.EventRunResumed,
		core.EventStepScheduled,
		core.EventStepStarted,
		core.EventStepCompleted,
		core.EventStepFailed,
		core.EventStepRetrying,
		core.EventStepSkipped,
		core.EventStepCancelled,
	}

	now := time.Now().UTC().Truncate(time.Millisecond)

	for _, et := range allEventTypes {
		t.Run(string(et), func(t *testing.T) {
			original := &core.Event{
				ID:        core.NewID("event"),
				Type:      et,
				TenantID:  "tenant-serialization-test",
				RunID:     core.NewID("run"),
				StepID:    "step-roundtrip",
				Timestamp: now,
				Payload:   json.RawMessage(`{"event_meta":"ok","type":"` + string(et) + `"}`),
			}

			// Encode
			encoded, err := SerializeEvent(original)
			if err != nil {
				t.Fatalf("failed serializing event %s: %v", et, err)
			}

			// Decode
			decoded, err := DeserializeEvent(encoded)
			if err != nil {
				t.Fatalf("failed deserializing event %s: %v", et, err)
			}

			// Deep comparison
			if decoded.ID != original.ID {
				t.Errorf("ID mismatch for %s: got %s, expected %s", et, decoded.ID, original.ID)
			}
			if decoded.Type != original.Type {
				t.Errorf("Type mismatch: got %s, expected %s", decoded.Type, original.Type)
			}
			if decoded.TenantID != original.TenantID {
				t.Errorf("TenantID mismatch: got %s, expected %s", decoded.TenantID, original.TenantID)
			}
			if decoded.RunID != original.RunID {
				t.Errorf("RunID mismatch: got %s, expected %s", decoded.RunID, original.RunID)
			}
			if decoded.StepID != original.StepID {
				t.Errorf("StepID mismatch: got %s, expected %s", decoded.StepID, original.StepID)
			}
			if !decoded.Timestamp.Equal(original.Timestamp) {
				t.Errorf("Timestamp mismatch: got %v, expected %v", decoded.Timestamp, original.Timestamp)
			}
			if !reflect.DeepEqual(decoded.Payload, original.Payload) {
				t.Errorf("Payload mismatch: got %s, expected %s", string(decoded.Payload), string(original.Payload))
			}
		})
	}
}
