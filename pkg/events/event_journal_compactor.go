package events

import (
	"context"
	"sync"
	"time"
)

// JournalCompactorRecord models an append-only event log item for compaction.
type JournalCompactorRecord struct {
	SequenceID uint64    `json:"sequence_id"`
	Topic      string    `json:"topic"`
	Payload    string    `json:"payload"`
	Timestamp  time.Time `json:"timestamp"`
}

// EventJournalCompactor manages in-memory event retention and snapshot compaction.
type EventJournalCompactor struct {
	mu        sync.RWMutex
	records   []JournalCompactorRecord
	maxRetain int
	retention time.Duration
}

// NewEventJournalCompactor creates an event journal compaction manager.
func NewEventJournalCompactor(maxRetain int, retentionWindow time.Duration) *EventJournalCompactor {
	if maxRetain <= 0 {
		maxRetain = 1000
	}
	return &EventJournalCompactor{
		records:   make([]JournalCompactorRecord, 0),
		maxRetain: maxRetain,
		retention: retentionWindow,
	}
}

// Append inserts a record to the journal.
func (c *EventJournalCompactor) Append(record JournalCompactorRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records = append(c.records, record)
}

// Compact purges records older than retention time or truncates beyond maxRetain.
func (c *EventJournalCompactor) Compact(ctx context.Context, now time.Time) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := now.Add(-c.retention)
	retained := make([]JournalCompactorRecord, 0, len(c.records))

	for _, rec := range c.records {
		if c.retention > 0 && rec.Timestamp.Before(cutoff) {
			continue
		}
		retained = append(retained, rec)
	}

	if len(retained) > c.maxRetain {
		excess := len(retained) - c.maxRetain
		retained = retained[excess:]
	}

	removed := len(c.records) - len(retained)
	c.records = retained
	return removed
}

// Count returns current number of retained records.
func (c *EventJournalCompactor) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.records)
}
