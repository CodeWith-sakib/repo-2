package events

import (
	"context"
	"testing"
	"time"
)

func TestEventJournalCompactor(t *testing.T) {
	compactor := NewEventJournalCompactor(5, 10*time.Minute)
	now := time.Now()

	for i := 1; i <= 10; i++ {
		compactor.Append(JournalCompactorRecord{
			SequenceID: uint64(i),
			Topic:      "telemetry",
			Payload:    "data",
			Timestamp:  now.Add(time.Duration(-15+i) * time.Minute),
		})
	}

	if compactor.Count() != 10 {
		t.Errorf("expected 10 records, got %d", compactor.Count())
	}
	removed := compactor.Compact(context.Background(), now)
	if removed <= 0 {
		t.Errorf("expected removed > 0, got %d", removed)
	}
	if compactor.Count() > 5 {
		t.Errorf("expected count <= 5, got %d", compactor.Count())
	}
}
