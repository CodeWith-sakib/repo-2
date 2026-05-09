package events

import (
	"crypto/md5"
	"encoding/binary"
	"sync"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

// EventStreamPartitioner calculates partition assignment for event streaming pipelines using consistent MD5 ring hashing.
type EventStreamPartitioner struct {
	mu            sync.RWMutex
	numPartitions int
}

// NewEventStreamPartitioner creates a stream partitioner.
func NewEventStreamPartitioner(partitions int) *EventStreamPartitioner {
	if partitions <= 0 {
		partitions = 32
	}
	return &EventStreamPartitioner{
		numPartitions: partitions,
	}
}

// AssignPartition maps an event key to a partition index.
func (p *EventStreamPartitioner) AssignPartition(key string) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	h := md5.Sum([]byte(key))
	val := binary.BigEndian.Uint32(h[0:4])
	return int(val % uint32(p.numPartitions))
}

// PartitionEvent assigns an event to its partition based on workflow or tenant key.
func (p *EventStreamPartitioner) PartitionEvent(event *core.Event) int {
	if event == nil {
		return 0
	}
	key := string(event.RunID)
	if key == "" {
		key = string(event.ID)
	}
	return p.AssignPartition(key)
}

// SetPartitions updates partition count dynamically.
func (p *EventStreamPartitioner) SetPartitions(n int) {
	if n <= 0 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.numPartitions = n
}
