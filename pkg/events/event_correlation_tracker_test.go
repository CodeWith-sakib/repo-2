package events

import (
	"testing"
)

func TestEventCorrelationTracker(t *testing.T) {
	tracker := NewEventCorrelationTracker()

	// e1 -> e2, e3; e3 -> e4
	tracker.TrackEvent("e1", "")
	tracker.TrackEvent("e2", "e1")
	tracker.TrackEvent("e3", "e1")
	tracker.TrackEvent("e4", "e3")

	count := tracker.GetDescendantCount("e1")
	if count != 3 {
		t.Fatalf("expected 3 descendants for e1, got %d", count)
	}

	countE3 := tracker.GetDescendantCount("e3")
	if countE3 != 1 {
		t.Errorf("expected 1 descendant for e3, got %d", countE3)
	}
}
