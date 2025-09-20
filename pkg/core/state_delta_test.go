package core

import (
	"reflect"
	"testing"
	"time"
)

func TestStateDiffEngine_ComputeDiffAndApply(t *testing.T) {
	engine := NewStateDiffEngine()

	orig := map[string]interface{}{
		"retries": 1,
		"status":  "pending",
		"owner":   "dev",
	}

	updated := map[string]interface{}{
		"retries": 2,
		"status":  "completed",
		"output":  "ok",
	}

	deltas := engine.ComputeDiff(orig, updated)
	if len(deltas) != 4 {
		t.Fatalf("expected 4 deltas, got %d", len(deltas))
	}

	state := map[string]interface{}{
		"retries": 1,
		"status":  "pending",
		"owner":   "dev",
	}
	if err := engine.Apply(state, deltas); err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	if !reflect.DeepEqual(state, updated) {
		t.Errorf("state mismatch after apply:\ngot  %+v\nwant %+v", state, updated)
	}
}

func TestStateDiffEngine_InvertPatch(t *testing.T) {
	engine := NewStateDiffEngine()

	orig := map[string]interface{}{
		"count":  10,
		"status": "init",
	}

	patch := DeltaPatch{
		PatchID:    "patch-101",
		WorkflowID: "wf-1",
		Version:    1,
		Timestamp:  time.Now(),
		Deltas: []StateDelta{
			{Op: OpReplace, Path: "/status", OldValue: "init", NewValue: "running"},
			{Op: OpAdd, Path: "/progress", OldValue: nil, NewValue: 50},
			{Op: OpRemove, Path: "/count", OldValue: 10, NewValue: nil},
		},
	}

	mutated := map[string]interface{}{
		"count":  10,
		"status": "init",
	}
	if err := engine.Apply(mutated, patch.Deltas); err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	invPatch := engine.InvertPatch(patch)
	if err := engine.Apply(mutated, invPatch.Deltas); err != nil {
		t.Fatalf("invert apply failed: %v", err)
	}

	if !reflect.DeepEqual(mutated, orig) {
		t.Errorf("state mismatch after invert apply:\ngot  %+v\nwant %+v", mutated, orig)
	}
}

func TestStateDiffEngine_CompressDeltas(t *testing.T) {
	engine := NewStateDiffEngine()

	deltas := []StateDelta{
		{Op: OpAdd, Path: "/temp", OldValue: nil, NewValue: "foo"},
		{Op: OpReplace, Path: "/temp", OldValue: "foo", NewValue: "bar"},
		{Op: OpRemove, Path: "/temp", OldValue: "bar", NewValue: nil}, // should cancel out completely
		{Op: OpReplace, Path: "/count", OldValue: 1, NewValue: 2},
		{Op: OpReplace, Path: "/count", OldValue: 2, NewValue: 3},
	}

	compressed := engine.CompressDeltas(deltas)
	if len(compressed) != 1 {
		t.Fatalf("expected 1 compressed delta, got %d: %+v", len(compressed), compressed)
	}
	if compressed[0].Path != "/count" || compressed[0].NewValue != 3 {
		t.Errorf("unexpected compressed result: %+v", compressed[0])
	}
}
