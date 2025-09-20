package core

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"
)

// DeltaOp represents an operation on a JSON state path.
type DeltaOp string

const (
	OpAdd     DeltaOp = "add"
	OpReplace DeltaOp = "replace"
	OpRemove  DeltaOp = "remove"
)

// StateDelta represents a single atomic state transition.
type StateDelta struct {
	Op       DeltaOp     `json:"op"`
	Path     string      `json:"path"`
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
}

// DeltaPatch is an ordered sequence of state deltas with transaction metadata.
type DeltaPatch struct {
	PatchID     string       `json:"patch_id"`
	WorkflowID  string       `json:"workflow_id"`
	Version     int64        `json:"version"`
	Timestamp   time.Time    `json:"timestamp"`
	Deltas      []StateDelta `json:"deltas"`
	Description string       `json:"description,omitempty"`
}

// StateDiffEngine computes, inverts, and applies state deltas for workflow executions.
type StateDiffEngine struct {
	mu sync.RWMutex
}

// NewStateDiffEngine creates an instance of the diff engine.
func NewStateDiffEngine() *StateDiffEngine {
	return &StateDiffEngine{}
}

// ComputeDiff calculates the deltas needed to transform original into updated.
func (e *StateDiffEngine) ComputeDiff(original, updated map[string]interface{}) []StateDelta {
	var deltas []StateDelta

	// Check keys in original
	for k, oldV := range original {
		newV, exists := updated[k]
		if !exists {
			deltas = append(deltas, StateDelta{
				Op:       OpRemove,
				Path:     "/" + k,
				OldValue: oldV,
				NewValue: nil,
			})
			continue
		}

		if !reflect.DeepEqual(oldV, newV) {
			deltas = append(deltas, StateDelta{
				Op:       OpReplace,
				Path:     "/" + k,
				OldValue: oldV,
				NewValue: newV,
			})
		}
	}

	// Check keys added in updated
	for k, newV := range updated {
		if _, exists := original[k]; !exists {
			deltas = append(deltas, StateDelta{
				Op:       OpAdd,
				Path:     "/" + k,
				OldValue: nil,
				NewValue: newV,
			})
		}
	}

	sort.Slice(deltas, func(i, j int) bool {
		return deltas[i].Path < deltas[j].Path
	})

	return deltas
}

// InvertPatch creates an inverse patch that rolls back changes made by patch.
func (e *StateDiffEngine) InvertPatch(patch DeltaPatch) DeltaPatch {
	invDeltas := make([]StateDelta, 0, len(patch.Deltas))

	// Invert in reverse order
	for i := len(patch.Deltas) - 1; i >= 0; i-- {
		d := patch.Deltas[i]
		switch d.Op {
		case OpAdd:
			invDeltas = append(invDeltas, StateDelta{
				Op:       OpRemove,
				Path:     d.Path,
				OldValue: d.NewValue,
				NewValue: nil,
			})
		case OpRemove:
			invDeltas = append(invDeltas, StateDelta{
				Op:       OpAdd,
				Path:     d.Path,
				OldValue: nil,
				NewValue: d.OldValue,
			})
		case OpReplace:
			invDeltas = append(invDeltas, StateDelta{
				Op:       OpReplace,
				Path:     d.Path,
				OldValue: d.NewValue,
				NewValue: d.OldValue,
			})
		}
	}

	return DeltaPatch{
		PatchID:     "inv-" + patch.PatchID,
		WorkflowID:  patch.WorkflowID,
		Version:     patch.Version,
		Timestamp:   time.Now(),
		Deltas:      invDeltas,
		Description: fmt.Sprintf("Inverse of patch %s", patch.PatchID),
	}
}

// Apply applies a slice of deltas onto a target state map, mutating it.
func (e *StateDiffEngine) Apply(state map[string]interface{}, deltas []StateDelta) error {
	for _, d := range deltas {
		key := d.Path
		if len(key) > 0 && key[0] == '/' {
			key = key[1:]
		}

		switch d.Op {
		case OpAdd, OpReplace:
			state[key] = d.NewValue
		case OpRemove:
			delete(state, key)
		default:
			return fmt.Errorf("unsupported delta op: %s", d.Op)
		}
	}
	return nil
}

// CompressDeltas collapses redundant intermediate modifications targeting the same path.
func (e *StateDiffEngine) CompressDeltas(deltas []StateDelta) []StateDelta {
	type entry struct {
		firstOp   DeltaOp
		lastOp    DeltaOp
		origOld   interface{}
		latestNew interface{}
	}

	paths := make(map[string]*entry)
	var order []string

	for _, d := range deltas {
		ent, ok := paths[d.Path]
		if !ok {
			paths[d.Path] = &entry{
				firstOp:   d.Op,
				lastOp:    d.Op,
				origOld:   d.OldValue,
				latestNew: d.NewValue,
			}
			order = append(order, d.Path)
		} else {
			ent.lastOp = d.Op
			ent.latestNew = d.NewValue
		}
	}

	var compressed []StateDelta
	for _, p := range order {
		ent := paths[p]
		if ent.firstOp == OpAdd && ent.lastOp == OpRemove {
			// Added and then removed -> no-op!
			continue
		}
		if ent.firstOp == OpAdd {
			compressed = append(compressed, StateDelta{
				Op:       OpAdd,
				Path:     p,
				OldValue: nil,
				NewValue: ent.latestNew,
			})
		} else if ent.lastOp == OpRemove {
			compressed = append(compressed, StateDelta{
				Op:       OpRemove,
				Path:     p,
				OldValue: ent.origOld,
				NewValue: nil,
			})
		} else {
			// Replace
			if reflect.DeepEqual(ent.origOld, ent.latestNew) {
				// Mutated and restored back to original -> no-op
				continue
			}
			compressed = append(compressed, StateDelta{
				Op:       OpReplace,
				Path:     p,
				OldValue: ent.origOld,
				NewValue: ent.latestNew,
			})
		}
	}

	return compressed
}

// SerializePatch returns standard JSON bytes for a DeltaPatch.
func SerializePatch(patch DeltaPatch) ([]byte, error) {
	return json.Marshal(patch)
}

// DeserializePatch parses JSON into a DeltaPatch.
func DeserializePatch(data []byte) (*DeltaPatch, error) {
	var patch DeltaPatch
	if err := json.Unmarshal(data, &patch); err != nil {
		return nil, fmt.Errorf("failed to unmarshal delta patch: %w", err)
	}
	return &patch, nil
}
