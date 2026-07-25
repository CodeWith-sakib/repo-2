package events

import (
	"sync"
	"time"
)

// PartitionLagStatus reports consumer group offsets vs producer high watermarks.
type PartitionLagStatus struct {
	PartitionID  int       `json:"partition_id"`
	HighWatermark int64    `json:"high_watermark"`
	CurrentOffset int64    `json:"current_offset"`
	LagRecords   int64     `json:"lag_records"`
	ObservedAt   time.Time `json:"observed_at"`
}

// ConsumerLagMonitor tracks consumer group latency across partitions.
type ConsumerLagMonitor struct {
	mu           sync.RWMutex
	partitionLag map[int]PartitionLagStatus
}

// NewConsumerLagMonitor creates a lag monitoring tracker.
func NewConsumerLagMonitor() *ConsumerLagMonitor {
	return &ConsumerLagMonitor{
		partitionLag: make(map[int]PartitionLagStatus),
	}
}

// RecordLag stores the observed offsets for a partition.
func (m *ConsumerLagMonitor) RecordLag(partitionID int, highWatermark, currentOffset int64) PartitionLagStatus {
	m.mu.Lock()
	defer m.mu.Unlock()

	lag := highWatermark - currentOffset
	if lag < 0 {
		lag = 0
	}

	status := PartitionLagStatus{
		PartitionID:   partitionID,
		HighWatermark: highWatermark,
		CurrentOffset: currentOffset,
		LagRecords:    lag,
		ObservedAt:    time.Now(),
	}

	m.partitionLag[partitionID] = status
	return status
}

// TotalLag sums lag across all active partitions.
func (m *ConsumerLagMonitor) TotalLag() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var total int64
	for _, status := range m.partitionLag {
		total += status.LagRecords
	}
	return total
}
