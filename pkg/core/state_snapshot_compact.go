package core

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// SnapshotDelta represents an incremental state delta with sequence ordering.
type SnapshotDelta struct {
	Sequence  int64                  `json:"sequence"`
	Timestamp time.Time              `json:"timestamp"`
	Mutations map[string]interface{} `json:"mutations"`
	Deletions []string               `json:"deletions"`
}

// CompactedSnapshot summarizes cumulative execution state up to sequence number.
type CompactedSnapshot struct {
	BaseSequence int64                  `json:"base_sequence"`
	Timestamp    time.Time              `json:"timestamp"`
	State        map[string]interface{} `json:"state"`
}

// StateSnapshotCompactor compacts sequence of historical deltas into minimal baseline snapshot.
type StateSnapshotCompactor struct {
	mu sync.RWMutex
}

// NewStateSnapshotCompactor creates a new state snapshot compactor.
func NewStateSnapshotCompactor() *StateSnapshotCompactor {
	return &StateSnapshotCompactor{}
}

// Compact merges a base snapshot with subsequent ordered deltas into a new canonical snapshot.
func (c *StateSnapshotCompactor) Compact(base *CompactedSnapshot, deltas []SnapshotDelta) (*CompactedSnapshot, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	resultState := make(map[string]interface{})
	var currentSeq int64 = 0

	if base != nil {
		currentSeq = base.BaseSequence
		for k, v := range base.State {
			resultState[k] = v
		}
	}

	for _, delta := range deltas {
		if delta.Sequence <= currentSeq {
			return nil, fmt.Errorf("out-of-order delta sequence %d; expected > %d", delta.Sequence, currentSeq)
		}
		currentSeq = delta.Sequence

		for k, v := range delta.Mutations {
			resultState[k] = v
		}

		for _, delKey := range delta.Deletions {
			delete(resultState, delKey)
		}
	}

	return &CompactedSnapshot{
		BaseSequence: currentSeq,
		Timestamp:    time.Now().UTC(),
		State:        resultState,
	}, nil
}

// Serialize serializes compacted snapshot to JSON bytes.
func (s *CompactedSnapshot) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

// DeserializeCompactedSnapshot deserializes snapshot from JSON.
func DeserializeCompactedSnapshot(data []byte) (*CompactedSnapshot, error) {
	var snap CompactedSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}
