package postgres

import (
	"context"
	"testing"
	"time"
)

func TestDBConnectionCircuitBreaker(t *testing.T) {
	cb := NewDBConnectionCircuitBreaker(2, 50*time.Millisecond)

	if cb.IsOpen() {
		t.Fatal("expected initially closed breaker")
	}

	now := time.Now().UTC()
	cb.RecordFailure(now)
	if cb.IsOpen() {
		t.Error("should not be open after 1 failure")
	}

	// 2nd failure trips breaker
	cb.RecordFailure(now)
	if !cb.IsOpen() {
		t.Error("expected open breaker after 2 failures")
	}

	err := cb.PingOrCheck(context.Background(), nil)
	if err == nil {
		t.Error("expected circuit breaker error, got nil")
	}

	// Wait cooldown -> allowed ping
	time.Sleep(60 * time.Millisecond)
	err = cb.PingOrCheck(context.Background(), nil)
	if err != nil {
		t.Errorf("expected success after cooldown, got: %v", err)
	}
}
