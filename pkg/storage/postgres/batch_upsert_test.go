package postgres

import (
	"strings"
	"testing"
)

func TestUpsertBatch_DoUpdate(t *testing.T) {
	cfg := BatchUpsertConfig{
		Table:            "workflow_instances",
		Columns:          []string{"id", "tenant_id", "status", "version"},
		ConflictColumns:  []string{"id"},
		Action:           ConflictDoUpdate,
		UpdateColumns:    []string{"status", "version"},
		WhereClause:      "workflow_instances.version < EXCLUDED.version",
		ReturningColumns: []string{"id", "status"},
	}

	batch, err := NewUpsertBatch(cfg)
	if err != nil {
		t.Fatalf("failed to create UpsertBatch: %v", err)
	}

	rows := [][]interface{}{
		{"wf-1", "tenant-a", "running", 2},
		{"wf-2", "tenant-b", "failed", 3},
	}

	sql, args, err := batch.BuildQuery(rows)
	if err != nil {
		t.Fatalf("build query failed: %v", err)
	}

	if len(args) != 8 {
		t.Errorf("expected 8 args, got %d", len(args))
	}

	if !strings.Contains(sql, "INSERT INTO workflow_instances (id, tenant_id, status, version)") {
		t.Errorf("unexpected INSERT clause: %s", sql)
	}
	if !strings.Contains(sql, "VALUES ($1, $2, $3, $4), ($5, $6, $7, $8)") {
		t.Errorf("unexpected VALUES clause: %s", sql)
	}
	if !strings.Contains(sql, "ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status, version = EXCLUDED.version") {
		t.Errorf("unexpected ON CONFLICT clause: %s", sql)
	}
	if !strings.Contains(sql, "WHERE workflow_instances.version < EXCLUDED.version") {
		t.Errorf("unexpected WHERE clause: %s", sql)
	}
	if !strings.Contains(sql, "RETURNING id, status") {
		t.Errorf("unexpected RETURNING clause: %s", sql)
	}
}

func TestUpsertBatch_DoNothing(t *testing.T) {
	cfg := BatchUpsertConfig{
		Table:           "event_log",
		Columns:         []string{"event_id", "payload"},
		ConflictColumns: []string{"event_id"},
		Action:          ConflictDoNothing,
	}

	batch, err := NewUpsertBatch(cfg)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	rows := [][]interface{}{
		{"evt-1", `{"foo":"bar"}`},
	}

	sql, args, err := batch.BuildQuery(rows)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
	if !strings.Contains(sql, "ON CONFLICT (event_id) DO NOTHING") {
		t.Errorf("expected DO NOTHING, got: %s", sql)
	}
}

func TestChunkRows(t *testing.T) {
	rows := [][]interface{}{
		{1}, {2}, {3}, {4}, {5},
	}

	chunks := ChunkRows(rows, 2)
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	if len(chunks[0]) != 2 || len(chunks[1]) != 2 || len(chunks[2]) != 1 {
		t.Errorf("unexpected chunk sizes: %+v", chunks)
	}
}
