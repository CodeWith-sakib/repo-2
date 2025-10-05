package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

// ReplayFilter configures which events are included in the replay.
type ReplayFilter struct {
	TenantID  string
	Topic     string
	StartTime time.Time
	EndTime   time.Time
}

// ReplayOptions configures execution behavior of an event replay job.
type ReplayOptions struct {
	SpeedMultiplier float64 // 0 = as fast as possible, 1.0 = real-time, 2.0 = 2x speed
	BatchSize       int
}

// ReplayProgress tracks progress of a replay job.
type ReplayProgress struct {
	TotalEvents    int
	ReplayedEvents int
	SkippedEvents  int
	StartedAt      time.Time
	CompletedAt    time.Time
	IsDone         bool
}

// EventSink is a consumer function that receives replayed events.
type EventSink func(ctx context.Context, event *core.Event) error

// EventReplayer coordinates historical event streams and replay pacing.
type EventReplayer struct {
	mu     sync.RWMutex
	events []*core.Event
}

// NewEventReplayer creates an empty event replayer.
func NewEventReplayer() *EventReplayer {
	return &EventReplayer{}
}

// IngestHistory appends historical events into the replay buffer.
func (r *EventReplayer) IngestHistory(events ...*core.Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, events...)
}

// Replay streams matching events into the sink respecting the configured pacing.
func (r *EventReplayer) Replay(ctx context.Context, filter ReplayFilter, opts ReplayOptions, sink EventSink) (*ReplayProgress, error) {
	r.mu.RLock()
	allEvents := make([]*core.Event, len(r.events))
	copy(allEvents, r.events)
	r.mu.RUnlock()

	progress := &ReplayProgress{
		StartedAt: time.Now(),
	}

	// 1. Filter events
	var matching []*core.Event
	for _, e := range allEvents {
		if filter.TenantID != "" && e.TenantID != filter.TenantID {
			progress.SkippedEvents++
			continue
		}
		if filter.Topic != "" && string(e.Type) != filter.Topic {
			progress.SkippedEvents++
			continue
		}
		if !filter.StartTime.IsZero() && e.Timestamp.Before(filter.StartTime) {
			progress.SkippedEvents++
			continue
		}
		if !filter.EndTime.IsZero() && e.Timestamp.After(filter.EndTime) {
			progress.SkippedEvents++
			continue
		}
		matching = append(matching, e)
	}

	progress.TotalEvents = len(matching)

	var prevTime time.Time
	for i, e := range matching {
		select {
		case <-ctx.Done():
			return progress, ctx.Err()
		default:
		}

		// Pacing calculation
		if opts.SpeedMultiplier > 0 && i > 0 && !prevTime.IsZero() && !e.Timestamp.IsZero() {
			diff := e.Timestamp.Sub(prevTime)
			if diff > 0 {
				scaledDelay := time.Duration(float64(diff) / opts.SpeedMultiplier)
				// Cap max sleep in tests
				if scaledDelay > 100*time.Millisecond {
					scaledDelay = 100 * time.Millisecond
				}
				time.Sleep(scaledDelay)
			}
		}
		prevTime = e.Timestamp

		if err := sink(ctx, e); err != nil {
			return progress, fmt.Errorf("sink error at event %s: %w", e.ID, err)
		}

		progress.ReplayedEvents++
	}

	progress.CompletedAt = time.Now()
	progress.IsDone = true
	return progress, nil
}
