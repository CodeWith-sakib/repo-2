package storage

import (
	"fmt"
	"sync"
	"time"
)

// WALSegmentMetadata holds summary info for a WAL segment eligible for compaction.
type WALSegmentMetadata struct {
	SegmentID   string
	StartOffset uint64
	EndOffset   uint64
	EntryCount  int
	CreatedAt   time.Time
	Checksum    uint32
	Compacted   bool
}

// WALCompactorPolicy defines when segments are eligible for compaction.
type WALCompactorPolicy struct {
	MaxSegmentAge     time.Duration
	MaxSegmentBytes   uint64
	MinSegmentsToKeep int
}

// WALCompactionResult is returned after a compaction run.
type WALCompactionResult struct {
	SegmentsScanned   int
	SegmentsCompacted int
	BytesReclaimed    uint64
	Duration          time.Duration
}

// WALCompactor manages lifecycle compaction of WAL segments: identifies stale segments,
// merges eligible entries, and removes redundant data to bound disk usage.
type WALCompactor struct {
	mu       sync.Mutex
	segments map[string]*WALSegmentMetadata
	policy   WALCompactorPolicy
}

// NewWALCompactor initializes a compactor with the given policy.
func NewWALCompactor(policy WALCompactorPolicy) *WALCompactor {
	return &WALCompactor{
		segments: make(map[string]*WALSegmentMetadata),
		policy:   policy,
	}
}

// RegisterSegment registers a WAL segment for tracking.
func (c *WALCompactor) RegisterSegment(seg WALSegmentMetadata) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.segments[seg.SegmentID] = &seg
}

// Compact runs the compaction pass: marks qualifying segments as compacted
// and returns a summary of what was reclaimed.
func (c *WALCompactor) Compact() (*WALCompactionResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	start := time.Now()
	result := &WALCompactionResult{}

	active := []*WALSegmentMetadata{}
	for _, seg := range c.segments {
		if !seg.Compacted {
			active = append(active, seg)
		}
	}

	result.SegmentsScanned = len(active)

	if len(active) <= c.policy.MinSegmentsToKeep {
		return result, nil
	}

	now := time.Now()
	for _, seg := range active {
		eligible := false
		reason := ""

		ageExceeded := now.Sub(seg.CreatedAt) > c.policy.MaxSegmentAge
		sizeExceeded := (seg.EndOffset - seg.StartOffset) > c.policy.MaxSegmentBytes

		if ageExceeded {
			eligible = true
			reason = "age"
		} else if sizeExceeded {
			eligible = true
			reason = "size"
		}

		if eligible {
			seg.Compacted = true
			bytes := seg.EndOffset - seg.StartOffset
			result.SegmentsCompacted++
			result.BytesReclaimed += bytes
			_ = reason
		}

		// Respect minimum retention
		remaining := result.SegmentsScanned - result.SegmentsCompacted
		if remaining <= c.policy.MinSegmentsToKeep {
			break
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// PendingCount returns the number of segments not yet compacted.
func (c *WALCompactor) PendingCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	count := 0
	for _, s := range c.segments {
		if !s.Compacted {
			count++
		}
	}
	return count
}

// Summary generates a diagnostic string of segment states.
func (c *WALCompactor) Summary() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return fmt.Sprintf("wal_compactor: total=%d policy={age=%s minKeep=%d}",
		len(c.segments), c.policy.MaxSegmentAge, c.policy.MinSegmentsToKeep)
}
