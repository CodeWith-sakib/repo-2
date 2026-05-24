package events

import (
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

// EventReplayCriteria defines matching criteria for replaying historical events.
type EventReplayCriteria struct {
	TenantID     string
	EventTypes   []string
	Since        time.Time
	Until        time.Time
	MaxBatchSize int
}

// EventReplayFilter matches events against replay criteria.
type EventReplayFilter struct {
	criteria EventReplayCriteria
}

// NewEventReplayFilter creates an event replay filter.
func NewEventReplayFilter(criteria EventReplayCriteria) *EventReplayFilter {
	return &EventReplayFilter{criteria: criteria}
}

// Matches checks if a candidate event satisfies all filter predicates.
func (f *EventReplayFilter) Matches(event *core.Event) bool {
	if event == nil {
		return false
	}

	if f.criteria.TenantID != "" && event.TenantID != f.criteria.TenantID {
		return false
	}

	if len(f.criteria.EventTypes) > 0 {
		matchedType := false
		for _, t := range f.criteria.EventTypes {
			if string(event.Type) == t {
				matchedType = true
				break
			}
		}
		if !matchedType {
			return false
		}
	}

	if !f.criteria.Since.IsZero() && event.Timestamp.Before(f.criteria.Since) {
		return false
	}

	if !f.criteria.Until.IsZero() && event.Timestamp.After(f.criteria.Until) {
		return false
	}

	return true
}
