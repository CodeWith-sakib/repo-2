package events

import (
	"testing"
	"time"
)

func TestEventPublishCircuitBreaker(t *testing.T) {
	cb := NewEventPublishCircuitBreaker(3, 50*time.Millisecond)

	now := time.Now().UTC()
	if !cb.CanPublish(now) {
		t.Fatal("expected initially closed breaker to allow publish")
	}

	cb.RecordFailure(now)
	cb.RecordFailure(now)
	if cb.IsOpen() {
		t.Error("breaker should not be open after 2 failures")
	}

	// 3rd failure trips breaker
	cb.RecordFailure(now)
	if !cb.IsOpen() {
		t.Error("breaker should be open after 3 failures")
	}

	if cb.CanPublish(now.Add(10 * time.Millisecond)) {
		t.Error("expected CanPublish to return false during cooldown")
	}

	// After cooldown -> allowed trial
	future := now.Add(60 * time.Millisecond)
	if !cb.CanPublish(future) {
		t.Error("expected trial allowed after cooldown")
	}

	cb.RecordSuccess()
	if cb.IsOpen() {
		t.Error("expected breaker closed after success")
	}
}
