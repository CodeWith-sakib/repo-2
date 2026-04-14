package events

import (
	"sync"
	"time"
)

// EventCorrelationChain maintains linked causal event relationship chains (cause -> effect).
type EventCorrelationChain struct {
	mu           sync.RWMutex
	parentToKids map[string][]string
	eventDetails map[string]time.Time
}

// NewEventCorrelationTracker creates a correlation tracker.
func NewEventCorrelationTracker() *EventCorrelationChain {
	return &EventCorrelationChain{
		parentToKids: make(map[string][]string),
		eventDetails: make(map[string]time.Time),
	}
}

// TrackEvent records an event and optional parent cause event ID.
func (c *EventCorrelationChain) TrackEvent(eventID, parentEventID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.eventDetails[eventID] = time.Now().UTC()
	if parentEventID != "" {
		c.parentToKids[parentEventID] = append(c.parentToKids[parentEventID], eventID)
	}
}

// GetDescendantCount recursively counts downstream events triggered by a root event.
func (c *EventCorrelationChain) GetDescendantCount(rootEventID string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	visited := make(map[string]bool)
	return c.countDescendants(rootEventID, visited)
}

func (c *EventCorrelationChain) countDescendants(id string, visited map[string]bool) int {
	kids, exists := c.parentToKids[id]
	if !exists {
		return 0
	}

	total := 0
	for _, k := range kids {
		if !visited[k] {
			visited[k] = true
			total += 1 + c.countDescendants(k, visited)
		}
	}
	return total
}
