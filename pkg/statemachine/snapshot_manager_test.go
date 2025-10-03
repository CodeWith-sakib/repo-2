package statemachine

import (
	"fmt"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func TestSnapshotManager_SaveAndGetLatest(t *testing.T) {
	mgr := NewSnapshotManager(3)

	runID := "run-100"
	for i := 1; i <= 5; i++ {
		snap := &WorkflowSnapshot{
			SnapshotID: fmt.Sprintf("snap-%d", i),
			RunID:      runID,
			WorkflowID: "wf-1",
			State:      core.RunStateRunning,
			Version:    int64(i),
			CreatedAt:  time.Now().Add(time.Duration(i) * time.Minute),
		}
		if err := mgr.SaveSnapshot(snap); err != nil {
			t.Fatalf("save snapshot %d failed: %v", i, err)
		}
	}

	// Should be pruned to 3
	snaps := mgr.ListSnapshots(runID)
	if len(snaps) != 3 {
		t.Fatalf("expected 3 snapshots retained, got %d", len(snaps))
	}

	latest, err := mgr.GetLatest(runID)
	if err != nil {
		t.Fatalf("get latest failed: %v", err)
	}
	if latest.Version != 5 {
		t.Errorf("expected latest version 5, got %d", latest.Version)
	}
}
