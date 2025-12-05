package worker

import (
	"context"
	"testing"
)

func TestCooperativeExecutionSession_ProgressAndResume(t *testing.T) {
	rec := NewMemoryCheckpointRecorder()
	session := NewCooperativeExecutionSession(context.Background(), "task-batch-import", rec)

	// Report 50% progress
	state := map[string]int{"rows_imported": 5000}
	if err := session.ReportProgress(50.0, "processing_chunks", state); err != nil {
		t.Fatalf("report progress failed: %v", err)
	}

	// Resume state from another session on same task ID
	session2 := NewCooperativeExecutionSession(context.Background(), "task-batch-import", rec)
	cp, ok := session2.ResumeState()
	if !ok || cp == nil {
		t.Fatal("expected to resume checkpoint")
	}

	if cp.ProgressPct != 50.0 || cp.Stage != "processing_chunks" {
		t.Errorf("unexpected checkpoint: %+v", cp)
	}
}

func TestCooperativeExecutionSession_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	rec := NewMemoryCheckpointRecorder()
	session := NewCooperativeExecutionSession(ctx, "task-cancelled", rec)

	cancel() // cancel context immediately

	err := session.ReportProgress(10.0, "start", nil)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
