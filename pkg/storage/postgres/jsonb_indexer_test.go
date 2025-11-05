package postgres

import (
	"strings"
	"testing"
)

func TestJSONBIndexBuilder_GINIndex(t *testing.T) {
	b, err := NewJSONBIndexBuilder("workflows", "payload")
	if err != nil {
		t.Fatalf("failed to create builder: %v", err)
	}

	ddl := b.CreateGINIndex("", GINPathOps)
	expected := "CREATE INDEX IF NOT EXISTS idx_workflows_payload_gin ON workflows USING gin (payload jsonb_path_ops);"
	if ddl != expected {
		t.Errorf("got %q, want %q", ddl, expected)
	}
}

func TestJSONBIndexBuilder_PathIndex(t *testing.T) {
	b, _ := NewJSONBIndexBuilder("events", "data")

	ddl := b.CreatePathIndex("", "user.id", "bigint")
	if !strings.Contains(ddl, "ON events (((data->'user'->>'id')::bigint));") {
		t.Errorf("unexpected path index DDL: %s", ddl)
	}
}

func TestJSONBIndexBuilder_ContainmentPredicate(t *testing.T) {
	b, _ := NewJSONBIndexBuilder("runs", "metadata")

	pred, err := b.BuildContainmentPredicate(map[string]interface{}{"status": "active", "priority": 10})
	if err != nil {
		t.Fatalf("containment predicate error: %v", err)
	}

	if !strings.Contains(pred, "metadata @>") || !strings.Contains(pred, "active") {
		t.Errorf("unexpected predicate: %s", pred)
	}
}
