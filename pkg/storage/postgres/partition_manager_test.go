package postgres

import (
	"strings"
	"testing"
	"time"
)

func TestTablePartitionManager_MonthlyPlanning(t *testing.T) {
	mgr, err := NewTablePartitionManager("events", "created_at", GranularityMonthly)
	if err != nil {
		t.Fatalf("create manager failed: %v", err)
	}

	start := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	plans := mgr.PlanPartitionsAhead(start, 3)

	if len(plans) != 3 {
		t.Fatalf("expected 3 plans, got %d", len(plans))
	}

	// First month: March 2026
	if plans[0].PartitionName != "events_y2026m03" {
		t.Errorf("expected events_y2026m03, got %s", plans[0].PartitionName)
	}
	if plans[1].PartitionName != "events_y2026m04" {
		t.Errorf("expected events_y2026m04, got %s", plans[1].PartitionName)
	}
	if plans[2].PartitionName != "events_y2026m05" {
		t.Errorf("expected events_y2026m05, got %s", plans[2].PartitionName)
	}

	ddl := mgr.CreatePartitionDDL(plans[0])
	if !strings.Contains(ddl, "CREATE TABLE IF NOT EXISTS events_y2026m03 PARTITION OF events") {
		t.Errorf("unexpected DDL: %s", ddl)
	}

	detach := mgr.DetachPartitionDDL("events_y2026m01", true)
	if detach != "ALTER TABLE events DETACH PARTITION events_y2026m01 CONCURRENTLY;" {
		t.Errorf("unexpected detach DDL: %s", detach)
	}
}

func TestTablePartitionManager_ParentDDL(t *testing.T) {
	mgr, _ := NewTablePartitionManager("audit_logs", "logged_at", GranularityDaily)
	ddl := mgr.ParentTableDDL("    id BIGSERIAL,\n    logged_at TIMESTAMP NOT NULL")

	if !strings.Contains(ddl, "PARTITION BY RANGE (logged_at)") {
		t.Errorf("missing partition by range: %s", ddl)
	}
}
