package postgres

import (
	"strings"
	"testing"
)

func TestAdvancedSelectBuilder_WithCTEAndWindow(t *testing.T) {
	b := NewAdvancedSelectBuilder()
	sql, args, err := b.
		WithCTE("recent_runs", "SELECT id, workflow_id, duration_ms FROM workflow_runs WHERE status = 'completed'").
		From("recent_runs").
		Select("id", "workflow_id", "ROW_NUMBER() OVER w as rank").
		Window("w", []string{"workflow_id"}, []string{"duration_ms DESC"}).
		Where("duration_ms > $1", 500).
		OrderBy("rank ASC").
		Limit(10).
		Build()

	if err != nil {
		t.Fatalf("unexpected error building query: %v", err)
	}

	if !strings.HasPrefix(sql, "WITH recent_runs AS (") {
		t.Errorf("expected query to start with CTE: %s", sql)
	}
	if !strings.Contains(sql, "WINDOW w AS (PARTITION BY workflow_id ORDER BY duration_ms DESC)") {
		t.Errorf("missing window definition: %s", sql)
	}
	if len(args) != 1 || args[0] != 500 {
		t.Errorf("expected argument 500, got %v", args)
	}
}
