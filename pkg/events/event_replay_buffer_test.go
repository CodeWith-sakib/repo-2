package events

import (
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestEventReplayBuffer(t *testing.T) {
	buf := NewEventReplayBuffer(3)

	seq1 := buf.Append(&core.Event{ID: "e1"})
	_ = buf.Append(&core.Event{ID: "e2"})
	seq3 := buf.Append(&core.Event{ID: "e3"})

	if seq3 != 3 {
		t.Fatalf("expected head seq 3, got %d", seq3)
	}

	// Replay after seq1 -> returns e2, e3
	evts := buf.ReplaySince(seq1)
	if len(evts) != 2 {
		t.Fatalf("expected 2 events in replay, got %d", len(evts))
	}
	if evts[0].ID != "e2" || evts[1].ID != "e3" {
		t.Errorf("unexpected replayed events: %v, %v", evts[0].ID, evts[1].ID)
	}

	// Append 4th -> evicts e1
	buf.Append(&core.Event{ID: "e4"})
	all := buf.ReplaySince(0)
	if len(all) != 3 {
		t.Fatalf("expected 3 events within capacity, got %d", len(all))
	}
	if all[0].ID != "e2" {
		t.Errorf("expected e2 to be oldest after eviction, got %s", all[0].ID)
	}
}
