package core

import (
	"testing"
	"time"
)

func TestParseCron(t *testing.T) {
	sched, err := ParseCron("*/15 2 1-5 * *")
	if err != nil {
		t.Fatalf("failed parsing cron: %v", err)
	}

	if len(sched.Minute) != 4 { // 0, 15, 30, 45
		t.Errorf("expected 4 minute values, got %d", len(sched.Minute))
	}
	if len(sched.Hour) != 1 || sched.Hour[0] != 2 {
		t.Errorf("expected hour 2, got %v", sched.Hour)
	}
	if len(sched.DayOfMonth) != 5 {
		t.Errorf("expected 5 dom values, got %d", len(sched.DayOfMonth))
	}

	// Negative cases
	badCases := []string{
		"* * *",
		"65 * * * *",
		"* 25 * * *",
		"invalid cron",
	}
	for _, bc := range badCases {
		if _, err := ParseCron(bc); err == nil {
			t.Errorf("expected error for bad cron %q, got nil", bc)
		}
	}
}

func TestCronNext(t *testing.T) {
	sched, _ := ParseCron("0 12 * * *") // Daily at noon
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	next := sched.Next(base)

	expected := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("next time mismatch: got %v, expected %v", next, expected)
	}
}
