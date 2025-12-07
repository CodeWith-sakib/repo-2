package core

import (
	"testing"
	"time"
)

func TestTemporalScheduler(t *testing.T) {
	ts := NewTemporalScheduler()

	// 1. Interval
	spec, err := ts.RegisterSchedule("heartbeat", ScheduleTypeInterval, "15s", "UTC")
	if err != nil {
		t.Fatalf("unexpected error registering interval: %v", err)
	}
	if spec.Interval != 15*time.Second {
		t.Errorf("expected 15s interval, got %v", spec.Interval)
	}

	now := time.Now().UTC()
	next, err := ts.NextOccurrence("heartbeat", now)
	if err != nil {
		t.Fatalf("unexpected error getting next occurrence: %v", err)
	}
	if next.Sub(now) < 14*time.Second {
		t.Errorf("next occurrence too early: %v", next)
	}

	// 2. Cron
	_, err = ts.RegisterSchedule("daily_sync", ScheduleTypeCron, "*/5 * * * *", "UTC")
	if err != nil {
		t.Fatalf("unexpected error registering cron: %v", err)
	}
	cronNext, err := ts.NextOccurrence("daily_sync", now)
	if err != nil {
		t.Fatalf("unexpected error calculating next cron: %v", err)
	}
	if cronNext.Minute()%5 != 0 {
		t.Errorf("expected minute divisible by 5, got %d", cronNext.Minute())
	}

	// 3. Once
	future := now.Add(2 * time.Hour).Format(time.RFC3339)
	_, err = ts.RegisterSchedule("one_off", ScheduleTypeOnce, future, "UTC")
	if err != nil {
		t.Fatalf("unexpected error registering once schedule: %v", err)
	}
	onceNext, err := ts.NextOccurrence("one_off", now)
	if err != nil {
		t.Fatalf("unexpected error calculating once next: %v", err)
	}
	if !onceNext.After(now) {
		t.Errorf("expected future time, got %v", onceNext)
	}

	if ts.Count() != 3 {
		t.Errorf("expected 3 registered schedules, got %d", ts.Count())
	}
}
