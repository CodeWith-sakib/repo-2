package events

import (
	"fmt"
	"hash/fnv"
	"sync"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

// PartitionedEvent wraps an event with an assigned partition and log sequence offset.
type PartitionedEvent struct {
	Event     *core.Event
	Partition int
	Offset    int64
}

// PartitionLog stores ordered events for a single partition.
type PartitionLog struct {
	mu     sync.RWMutex
	id     int
	events []*PartitionedEvent
}

// PartitionedTopic manages multiple log partitions and consumer group offsets for a topic.
type PartitionedTopic struct {
	mu             sync.RWMutex
	name           string
	numPartitions  int
	partitions     []*PartitionLog
	groupOffsets   map[string]map[int]int64 // groupName -> partitionID -> committedOffset
	roundRobinNext int
}

// NewPartitionedTopic creates a topic divided into numPartitions.
func NewPartitionedTopic(name string, numPartitions int) (*PartitionedTopic, error) {
	if name == "" || numPartitions <= 0 {
		return nil, fmt.Errorf("invalid topic name %q or partition count %d", name, numPartitions)
	}

	partitions := make([]*PartitionLog, numPartitions)
	for i := 0; i < numPartitions; i++ {
		partitions[i] = &PartitionLog{id: i}
	}

	return &PartitionedTopic{
		name:          name,
		numPartitions: numPartitions,
		partitions:    partitions,
		groupOffsets:  make(map[string]map[int]int64),
	}, nil
}

// Publish routes and appends an event to an appropriate partition based on routing key.
func (t *PartitionedTopic) Publish(event *core.Event, routingKey string) (*PartitionedEvent, error) {
	if event == nil {
		return nil, fmt.Errorf("nil event")
	}

	t.mu.Lock()
	var partID int
	if routingKey != "" {
		h := fnv.New32a()
		_, _ = h.Write([]byte(routingKey))
		partID = int(h.Sum32()) % t.numPartitions
		if partID < 0 {
			partID = -partID
		}
	} else {
		partID = t.roundRobinNext % t.numPartitions
		t.roundRobinNext++
	}
	t.mu.Unlock()

	pLog := t.partitions[partID]
	pLog.mu.Lock()
	defer pLog.mu.Unlock()

	offset := int64(len(pLog.events))
	pe := &PartitionedEvent{
		Event:     event,
		Partition: partID,
		Offset:    offset,
	}
	pLog.events = append(pLog.events, pe)

	return pe, nil
}

// ReadFromOffset reads up to maxCount events from partition starting at offset.
func (t *PartitionedTopic) ReadFromOffset(partitionID int, fromOffset int64, maxCount int) ([]*PartitionedEvent, error) {
	if partitionID < 0 || partitionID >= t.numPartitions {
		return nil, fmt.Errorf("invalid partition ID: %d", partitionID)
	}
	if maxCount <= 0 {
		maxCount = 100
	}

	pLog := t.partitions[partitionID]
	pLog.mu.RLock()
	defer pLog.mu.RUnlock()

	if fromOffset < 0 {
		fromOffset = 0
	}
	if fromOffset >= int64(len(pLog.events)) {
		return nil, nil
	}

	end := int(fromOffset) + maxCount
	if end > len(pLog.events) {
		end = len(pLog.events)
	}

	res := make([]*PartitionedEvent, end-int(fromOffset))
	copy(res, pLog.events[fromOffset:end])
	return res, nil
}

// CommitOffset marks progress for a consumer group on a specific partition.
func (t *PartitionedTopic) CommitOffset(group string, partitionID int, offset int64) error {
	if partitionID < 0 || partitionID >= t.numPartitions {
		return fmt.Errorf("invalid partition ID: %d", partitionID)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.groupOffsets[group]; !ok {
		t.groupOffsets[group] = make(map[int]int64)
	}
	t.groupOffsets[group][partitionID] = offset
	return nil
}

// GetCommittedOffset returns the last committed offset for a group/partition.
func (t *PartitionedTopic) GetCommittedOffset(group string, partitionID int) int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if m, ok := t.groupOffsets[group]; ok {
		if off, exists := m[partitionID]; exists {
			return off
		}
	}
	return 0
}
