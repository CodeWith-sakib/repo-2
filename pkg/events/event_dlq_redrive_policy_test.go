package events

import (
	"context"
	"errors"
	"testing"
)

func TestDLQRedriveManager(t *testing.T) {
	mgr := NewDLQRedriveManager(RedrivePolicy{
		MaxRedrives:   2,
		FallbackTopic: "events.poison.archive",
	})

	if !mgr.CanRedrive("evt-1") {
		t.Fatal("should be able to redrive initially")
	}

	c1, err := mgr.RecordRedriveAttempt(context.Background(), "evt-1")
	if err != nil || c1 != 1 {
		t.Fatalf("first redrive failed: %v", err)
	}

	c2, err := mgr.RecordRedriveAttempt(context.Background(), "evt-1")
	if err != nil || c2 != 2 {
		t.Fatalf("second redrive failed: %v", err)
	}

	// 3rd attempt should fail
	_, err = mgr.RecordRedriveAttempt(context.Background(), "evt-1")
	if !errors.Is(err, ErrMaxRedriveExceeded) {
		t.Fatalf("expected ErrMaxRedriveExceeded, got %v", err)
	}

	topic := mgr.TargetTopic("evt-1", "events.telemetry")
	if topic != "events.poison.archive" {
		t.Errorf("expected fallback topic, got %s", topic)
	}
}
