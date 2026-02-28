package events

import (
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

// BufferedEventRecord captures an event with sequence index and ingestion timestamp.
type BufferedEventRecord struct {
	Sequence   int64       `json:"sequence"`
	Event      *core.Event `json:"event"`
	ReceivedAt time.Time   `json:"received_at"`
}

// EventReplayBuffer maintains an in-memory chronological ring buffer for client reconnect replay.
type EventReplayBuffer struct {
	mu          sync.RWMutex
	records     []BufferedEventRecord
	capacity    int
	lastSeq     int64
}

// NewEventReplayBuffer creates a replay buffer with fixed item capacity.
func NewEventReplayBuffer(capacity int) *EventReplayBuffer {
	if capacity <= 0 {
		capacity = 5000
	}
	return &EventReplayBuffer{
		records:  make([]BufferedEventRecord, 0, capacity),
		capacity: capacity,
	}
}

// Append inserts a new event and returns its monotonic sequence number.
func (b *EventReplayBuffer) Append(evt *core.Event) int64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.lastSeq++
	rec := BufferedEventRecord{
		Sequence:   b.lastSeq,
		Event:      evt,
		ReceivedAt: time.Now().UTC(),
	}

	if len(b.records) >= b.capacity {
		// Drop oldest record
		b.records = b.records[1:]
	}
	b.records = append(b.records, rec)
	return b.lastSeq
}

// ReplaySince retrieves all events after the given sequence number.
func (b *EventReplayBuffer) ReplaySince(sinceSeq int64) []*core.Event {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var result []*core.Event
	for _, rec := range b.records {
		if rec.Sequence > sinceSeq {
			result = append(result, rec.Event)
		}
	}
	return result
}

// HeadSequence returns the highest sequence number in the buffer.
func (b *EventReplayBuffer) HeadSequence() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.lastSeq
}
