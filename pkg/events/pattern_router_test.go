package events

import (
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestPatternRouter_DispatchPriority(t *testing.T) {
	router := NewPatternRouter()

	var results []string

	_, _ = router.RegisterRoute(`^workflow\.`, "", 10, func(e *core.Event) error {
		results = append(results, "low")
		return nil
	})
	_, _ = router.RegisterRoute(`^workflow\.completed`, "", 20, func(e *core.Event) error {
		results = append(results, "high")
		return nil
	})

	ev := &core.Event{
		ID:        core.NewID("evt"),
		Type:      "workflow.completed",
		TenantID:  "tenant-x",
		Timestamp: time.Now(),
	}

	if err := router.Route(ev); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if len(results) != 1 || results[0] != "high" {
		t.Errorf("expected high-priority handler only, got %v", results)
	}
}

func TestPatternRouter_NoMatchError(t *testing.T) {
	router := NewPatternRouter()
	_, _ = router.RegisterRoute(`^task\.`, "corp", 5, func(_ *core.Event) error { return nil })

	ev := &core.Event{
		ID:        core.NewID("evt"),
		Type:      "storage.purge",
		TenantID:  "corp",
		Timestamp: time.Now(),
	}
	if err := router.Route(ev); err == nil {
		t.Error("expected no-route error")
	}
}
