package events

import (
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestEventReplayFilter(t *testing.T) {
	now := time.Now().UTC()
	crit := EventReplayCriteria{
		TenantID:   "tenant-corp",
		EventTypes: []string{"run.started", "run.completed"},
		Since:      now.Add(-1 * time.Hour),
		Until:      now.Add(1 * time.Hour),
	}

	filter := NewEventReplayFilter(crit)

	matchingEvt := &core.Event{
		TenantID:  "tenant-corp",
		Type:      "run.started",
		Timestamp: now,
	}
	if !filter.Matches(matchingEvt) {
		t.Error("expected matching event to pass filter")
	}

	mismatchTenant := &core.Event{
		TenantID:  "other-corp",
		Type:      "run.started",
		Timestamp: now,
	}
	if filter.Matches(mismatchTenant) {
		t.Error("expected mismatch tenant to fail filter")
	}

	mismatchType := &core.Event{
		TenantID:  "tenant-corp",
		Type:      "run.failed",
		Timestamp: now,
	}
	if filter.Matches(mismatchType) {
		t.Error("expected mismatch type to fail filter")
	}
}
