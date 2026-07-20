package events

import (
	"testing"
	"time"
)

func TestWatermarkClockTracker(t *testing.T) {
	tracker := NewWatermarkClockTracker(1 * time.Second)

	now := time.Now()
	t1 := now.Add(-5 * time.Second)
	t2 := now.Add(-3 * time.Second)

	tracker.UpdatePartition(0, 100, t1)
	tracker.UpdatePartition(1, 250, t2)

	// Low watermark should be min(t1, t2) - 1s = t1 - 1s
	lw := tracker.LowWatermark()
	expected := t1.Add(-1 * time.Second)

	if !lw.Equal(expected) {
		t.Errorf("expected low watermark %v, got %v", expected, lw)
	}
}
