package events

import (
	"bytes"
	"testing"
)

func TestMemoryJournalIntegrity(t *testing.T) {
	j := NewMemoryJournal()

	payload1 := []byte(`{"event":"step.scheduled","step_id":"step_1"}`)
	rec1, err := j.Append(payload1)
	if err != nil {
		t.Fatalf("append failed: %v", err)
	}

	payload2 := []byte(`{"event":"step.completed","step_id":"step_1"}`)
	_, err = j.Append(payload2)
	if err != nil {
		t.Fatalf("append failed: %v", err)
	}

	records, err := j.ReadFrom(1, 10)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// Test binary serialization
	var buf bytes.Buffer
	if err := EncodeRecord(&buf, rec1); err != nil {
		t.Fatalf("encode record failed: %v", err)
	}
	if buf.Len() < 24+len(payload1) {
		t.Errorf("buffer shorter than expected header+payload")
	}
}
