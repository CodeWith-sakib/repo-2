package events

import (
	"strings"
)

type EventFilter struct {
	allowedPrefixes []string
}

func NewEventFilter(prefixes []string) *EventFilter {
	return &EventFilter{allowedPrefixes: prefixes}
}

func (f *EventFilter) Allows(eventType string) bool {
	if len(f.allowedPrefixes) == 0 {
		return true
	}
	for _, p := range f.allowedPrefixes {
		if strings.HasPrefix(eventType, p) {
			return true
		}
	}
	return false
}
