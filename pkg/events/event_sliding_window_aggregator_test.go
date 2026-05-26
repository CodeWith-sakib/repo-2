package events

import (
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestEventRateAggregator(t *testing.T) {
	agg := NewEventRateAggregator(10 * time.Second)

	now := time.Now().UTC()
	for i := 0; i < 50; i++ {
		agg.RecordTimestamp(now.Add(time.Duration(i*100) * time.Millisecond))
	}

	rate := agg.CurrentRatePerSec(now.Add(5 * time.Second))
	if rate <= 0 {
		t.Errorf("expected positive event rate, got %f", rate)
	}

	agg.RecordEvent(&core.Event{ID: "e1"})
}
