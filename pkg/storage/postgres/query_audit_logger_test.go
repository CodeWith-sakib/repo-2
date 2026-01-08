package postgres

import (
	"context"
	"testing"
	"time"
)

func TestQueryAuditLogger(t *testing.T) {
	logger := NewQueryAuditLogger(nil, 3)

	ctx := context.Background()
	logger.Record(ctx, QueryAuditRecord{QueryID: "q1", SQL: "SELECT 1", Duration: 10 * time.Millisecond, ExecutedAt: time.Now().UTC()})
	logger.Record(ctx, QueryAuditRecord{QueryID: "q2", SQL: "SELECT 2", Duration: 20 * time.Millisecond, ExecutedAt: time.Now().UTC()})
	logger.Record(ctx, QueryAuditRecord{QueryID: "q3", SQL: "SELECT 3", Duration: 30 * time.Millisecond, ExecutedAt: time.Now().UTC()})

	if logger.Count() != 3 {
		t.Fatalf("expected 3 records, got %d", logger.Count())
	}

	// Record 4th -> should evict q1
	logger.Record(ctx, QueryAuditRecord{QueryID: "q4", SQL: "SELECT 4", Duration: 40 * time.Millisecond, ExecutedAt: time.Now().UTC()})

	recs := logger.RecentRecords(3)
	if len(recs) != 3 {
		t.Fatalf("expected 3 records returned, got %d", len(recs))
	}
	if recs[0].QueryID != "q2" {
		t.Errorf("expected oldest to be q2, got %s", recs[0].QueryID)
	}
	if recs[2].QueryID != "q4" {
		t.Errorf("expected newest to be q4, got %s", recs[2].QueryID)
	}
}
