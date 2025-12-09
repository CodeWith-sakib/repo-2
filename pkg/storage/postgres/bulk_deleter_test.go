package postgres

import (
	"context"
	"testing"
	"time"
)

func TestBulkDeleterDryRun(t *testing.T) {
	bd := NewBulkDeleter(nil)
	opts := BulkDeleteOptions{
		TableName:    "workflow_audit_logs",
		WhereClause:  "created_at < NOW() - INTERVAL '90 days'",
		BatchSize:    100,
		SleepBetween: 1 * time.Millisecond,
		DryRun:       true,
	}

	res, err := bd.DeleteInBatches(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error in dry run: %v", err)
	}
	if res.TotalDeleted != 0 {
		t.Errorf("expected 0 deleted in dry run, got %d", res.TotalDeleted)
	}

	// Validate required where clause
	invalidOpts := BulkDeleteOptions{
		TableName: "workflow_audit_logs",
	}
	_, err = bd.DeleteInBatches(context.Background(), invalidOpts)
	if err == nil {
		t.Error("expected error with missing where clause, got nil")
	}
}
