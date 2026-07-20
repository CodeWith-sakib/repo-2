package events

import (
	"sync"
	"time"
)

// PartitionWatermark models the progress of an event stream partition.
type PartitionWatermark struct {
	PartitionID   int       `json:"partition_id"`
	CurrentOffset int64     `json:"current_offset"`
	EventTime     time.Time `json:"event_time"`
	WallClock     time.Time `json:"wall_clock"`
}

// WatermarkClockTracker computes the low watermark across distributed streams to safely trigger windows.
type WatermarkClockTracker struct {
	mu           sync.RWMutex
	maxDelay     time.Duration
	partitionMap map[int]PartitionWatermark
}

// NewWatermarkClockTracker creates a tracker with allowed out-of-orderness delay.
func NewWatermarkClockTracker(maxOutOfOrderDelay time.Duration) *WatermarkClockTracker {
	if maxOutOfOrderDelay <= 0 {
		maxOutOfOrderDelay = 2 * time.Second
	}
	return &WatermarkClockTracker{
		maxDelay:     maxOutOfOrderDelay,
		partitionMap: make(map[int]PartitionWatermark),
	}
}

// UpdatePartition records the latest event timestamp for a partition.
func (t *WatermarkClockTracker) UpdatePartition(partitionID int, offset int64, eventTime time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.partitionMap[partitionID] = PartitionWatermark{
		PartitionID:   partitionID,
		CurrentOffset: offset,
		EventTime:     eventTime,
		WallClock:     time.Now(),
	}
}

// LowWatermark computes the minimum event time across all tracked partitions minus max out-of-order delay.
func (t *WatermarkClockTracker) LowWatermark() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.partitionMap) == 0 {
		return time.Time{}
	}

	var minTime time.Time
	first := true

	for _, p := range t.partitionMap {
		if first || p.EventTime.Before(minTime) {
			minTime = p.EventTime
			first = false
		}
	}

	return minTime.Add(-t.maxDelay)
}
