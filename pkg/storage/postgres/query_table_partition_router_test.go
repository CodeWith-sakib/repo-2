package postgres

import (
	"strings"
	"testing"
	"time"
)

func TestTablePartitionRouter(t *testing.T) {
	router := NewTablePartitionRouter("kf_step_events", 7)

	ref := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)
	partName := router.PartitionForTime(ref)

	if !strings.HasPrefix(partName, "kf_step_events_y2026_w") {
		t.Errorf("unexpected partition name: %s", partName)
	}

	ddl := router.GenerateCreatePartitionSQL(ref)
	if !strings.Contains(ddl, "CREATE TABLE IF NOT EXISTS") || !strings.Contains(ddl, "PARTITION OF kf_step_events") {
		t.Errorf("unexpected DDL generated: %s", ddl)
	}
}
