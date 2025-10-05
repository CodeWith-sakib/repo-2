package events

import (
	"context"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestEventReplayer_FilterAndReplay(t *testing.T) {
	replayer := NewEventReplayer()

	baseTime := time.Now().Add(-10 * time.Minute)
	var history []*core.Event
	for i := 0; i < 10; i++ {
		tenant := "tenant-a"
		if i%2 == 1 {
			tenant = "tenant-b"
		}
		topic := "task.completed"
		if i%3 == 0 {
			topic = "workflow.failed"
		}
		history = append(history, &core.Event{
			ID:        core.NewID("evt"),
			Type:      core.EventType(topic),
			TenantID:  tenant,
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
		})
	}

	replayer.IngestHistory(history...)

	// Filter: only tenant-a and task.completed
	filter := ReplayFilter{
		TenantID: "tenant-a",
		Topic:    "task.completed",
	}

	var replayed []*core.Event
	sink := func(ctx context.Context, e *core.Event) error {
		replayed = append(replayed, e)
		return nil
	}

	progress, err := replayer.Replay(context.Background(), filter, ReplayOptions{SpeedMultiplier: 0}, sink)
	if err != nil {
		t.Fatalf("replay error: %v", err)
	}

	if !progress.IsDone {
		t.Error("expected progress.IsDone to be true")
	}

	for _, e := range replayed {
		if e.TenantID != "tenant-a" || string(e.Type) != "task.completed" {
			t.Errorf("unexpected event in replay stream: %+v", e)
		}
	}
}
