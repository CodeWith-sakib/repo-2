package events

import (
	"testing"
	"time"
)

func TestDeduplicationWindow(t *testing.T) {
	window := NewDeduplicationWindow(100 * time.Millisecond)

	msg1 := []byte("event-payload-alpha")
	msg2 := []byte("event-payload-beta")

	if window.IsDuplicate(msg1) {
		t.Errorf("first appearance of msg1 should not be duplicate")
	}
	if !window.IsDuplicate(msg1) {
		t.Errorf("second appearance of msg1 within window must be duplicate")
	}

	if window.IsDuplicate(msg2) {
		t.Errorf("first appearance of msg2 should not be duplicate")
	}

	time.Sleep(120 * time.Millisecond)
	// Prune
	pruned := window.PruneExpired(time.Now())
	if pruned != 2 {
		t.Errorf("expected 2 keys pruned, got %d", pruned)
	}

	// Now msg1 should not be duplicate
	if window.IsDuplicate(msg1) {
		t.Errorf("msg1 should not be duplicate after expiration")
	}
}
