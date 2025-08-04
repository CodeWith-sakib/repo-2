package events

import (
	"sync"
	"time"
)

type DeadLetterItem struct {
	EventID      string
	Payload      []byte
	FailureReason string
	FailedAt     time.Time
	Attempts     int
}

type DeadLetterQueue struct {
	mu    sync.RWMutex
	items []*DeadLetterItem
}

func NewDeadLetterQueue() *DeadLetterQueue {
	return &DeadLetterQueue{
		items: make([]*DeadLetterItem, 0),
	}
}

func (d *DeadLetterQueue) Enqueue(eventID string, payload []byte, reason string, attempts int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.items = append(d.items, &DeadLetterItem{
		EventID:       eventID,
		Payload:       append([]byte(nil), payload...),
		FailureReason: reason,
		FailedAt:      time.Now().UTC(),
		Attempts:      attempts,
	})
}

func (d *DeadLetterQueue) Items() []*DeadLetterItem {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]*DeadLetterItem, len(d.items))
	copy(out, d.items)
	return out
}

func (d *DeadLetterQueue) Count() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.items)
}
