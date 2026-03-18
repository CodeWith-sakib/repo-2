package events

import (
	"fmt"
	"hash/fnv"
	"sync"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

// PriorityPartitionRouter maps events across partitioned Kafka/event topics based on tenant and priority.
type PriorityPartitionRouter struct {
	numPartitions int
	mu            sync.RWMutex
}

// NewPriorityPartitionRouter creates a priority-aware partition router.
func NewPriorityPartitionRouter(numPartitions int) *PriorityPartitionRouter {
	if numPartitions <= 0 {
		numPartitions = 16
	}
	return &PriorityPartitionRouter{
		numPartitions: numPartitions,
	}
}

// RoutePartition derives deterministic partition index for an event using FNV-1a hashing of key.
func (r *PriorityPartitionRouter) RoutePartition(tenantID string, eventKey string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(tenantID))
	_, _ = h.Write([]byte(":"))
	_, _ = h.Write([]byte(eventKey))

	return int(h.Sum32() % uint32(r.numPartitions))
}

// TargetTopic constructs qualified topic name with tier priority postfix.
func (r *PriorityPartitionRouter) TargetTopic(baseTopic string, priority string) string {
	if priority == "" {
		priority = "normal"
	}
	return fmt.Sprintf("%s.%s", baseTopic, priority)
}

// PartitionCount returns total partitions configured.
func (r *PriorityPartitionRouter) PartitionCount() int {
	return r.numPartitions
}

// Unused dummy to reference core.Event safely
func (r *PriorityPartitionRouter) ValidateEvent(evt *core.Event) bool {
	return evt != nil && evt.ID != ""
}
