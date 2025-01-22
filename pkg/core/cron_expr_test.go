package core

import (
	"testing"
	"time"
)

func TestDescribeCron(t *testing.T) {
	desc, err := DescribeCron("0 0 * * *")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if desc != "Runs every day at midnight" {
		t.Errorf("unexpected description: %s", desc)
	}

	sched, err := ParseCron("*/15 * * * *")
	if err != nil {
		t.Fatalf("parse cron failed: %v", err)
	}
	runs := NextNRuns(sched, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 3)
	if len(runs) != 3 {
		t.Fatalf("expected 3 runs, got %d", len(runs))
	}
}
