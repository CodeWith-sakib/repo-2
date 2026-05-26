package events

import (
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

// EventRateAggregator computes rolling event rates over a sliding window duration.
type EventRateAggregator struct {
	mu         sync.Mutex
	windowSize time.Duration
	timestamps []time.Time
}

// NewEventRateAggregator creates a sliding window rate aggregator.
func NewEventRateAggregator(windowSize time.Duration) *EventRateAggregator {
	if windowSize <= 0 {
		windowSize = 1 * time.Minute
	}
	return &EventRateAggregator{
		windowSize: windowSize,
		timestamps: make([]time.Time, 0, 100),
	}
}

// RecordEvent logs an incoming event timestamp into the sliding window.
func (a *EventRateAggregator) RecordEvent(event *core.Event) {
	if event == nil {
		return
	}
	a.RecordTimestamp(time.Now().UTC())
}

// RecordTimestamp logs an explicit event timestamp.
func (a *EventRateAggregator) RecordTimestamp(ts time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()

	cutoff := ts.Add(-a.windowSize)
	// Prune older than window
	validIdx := 0
	for i, t := range a.timestamps {
		if t.After(cutoff) {
			validIdx = i
			break
		}
		if i == len(a.timestamps)-1 {
			validIdx = len(a.timestamps)
		}
	}
	if validIdx > 0 {
		a.timestamps = a.timestamps[validIdx:]
	}

	a.timestamps = append(a.timestamps, ts)
}

// CurrentRatePerSec calculates average events per second across active sliding window.
func (a *EventRateAggregator) CurrentRatePerSec(now time.Time) float64 {
	a.mu.Lock()
	defer a.mu.Unlock()

	cutoff := now.Add(-a.windowSize)
	count := 0
	for _, t := range a.timestamps {
		if t.After(cutoff) {
			count++
		}
	}

	secs := a.windowSize.Seconds()
	if secs <= 0 {
		return 0
	}
	return float64(count) / secs
}
