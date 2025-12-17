package core

import (
	"testing"
	"time"
)

func TestStateSnapshotCompactor(t *testing.T) {
	compactor := NewStateSnapshotCompactor()

	base := &CompactedSnapshot{
		BaseSequence: 10,
		Timestamp:    time.Now().UTC(),
		State: map[string]interface{}{
			"retry_count": 1,
			"phase":       "INITIALIZING",
			"temp_flag":   true,
		},
	}

	deltas := []SnapshotDelta{
		{
			Sequence:  11,
			Timestamp: time.Now().UTC(),
			Mutations: map[string]interface{}{
				"phase": "RUNNING",
				"step":  "step_2",
			},
			Deletions: []string{"temp_flag"},
		},
		{
			Sequence:  12,
			Timestamp: time.Now().UTC(),
			Mutations: map[string]interface{}{
				"retry_count": 2,
				"finished":    true,
			},
			Deletions: nil,
		},
	}

	res, err := compactor.Compact(base, deltas)
	if err != nil {
		t.Fatalf("unexpected compaction error: %v", err)
	}

	if res.BaseSequence != 12 {
		t.Errorf("expected base sequence 12, got %d", res.BaseSequence)
	}

	if _, exists := res.State["temp_flag"]; exists {
		t.Error("expected temp_flag to be deleted")
	}

	if res.State["phase"] != "RUNNING" {
		t.Errorf("expected phase RUNNING, got %v", res.State["phase"])
	}

	// Verify serialization roundtrip
	data, err := res.Serialize()
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	deserialized, err := DeserializeCompactedSnapshot(data)
	if err != nil {
		t.Fatalf("failed to deserialize: %v", err)
	}

	if deserialized.BaseSequence != 12 {
		t.Errorf("deserialized base seq mismatch: %d", deserialized.BaseSequence)
	}

	// Verify out-of-order rejection
	badDeltas := []SnapshotDelta{
		{
			Sequence: 5,
		},
	}
	_, err = compactor.Compact(res, badDeltas)
	if err == nil {
		t.Error("expected error for out of order delta, got nil")
	}
}
