package postgres

import (
	"strings"
	"testing"
)

func TestQueryBuilderSelect(t *testing.T) {
	qb := NewQueryBuilder(DialectPostgreSQL)
	qb.Table("workflows").
		Select("id", "name", "version").
		Where("tenant_id = ?", "tenant-1").
		And("version >= ?", 2).
		OrderBy("created_at", OrderDesc).
		Limit(10).
		Offset(20).
		ForUpdate(true)

	sql, args := qb.Build()
	if !strings.Contains(sql, "SELECT id, name, version FROM workflows") {
		t.Errorf("unexpected SELECT clause: %s", sql)
	}
	if !strings.Contains(sql, "WHERE tenant_id = $1 AND version >= $2") {
		t.Errorf("unexpected WHERE clause: %s", sql)
	}
	if !strings.Contains(sql, "FOR UPDATE SKIP LOCKED") {
		t.Errorf("expected FOR UPDATE SKIP LOCKED, got %s", sql)
	}
	if len(args) != 2 || args[0] != "tenant-1" || args[1] != 2 {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestInsertBuilder(t *testing.T) {
	ib := NewInsertBuilder(DialectPostgreSQL)
	ib.Into("workflows").
		Columns("id", "name", "version").
		Values("wf-1", "pipeline", 1).
		Values("wf-2", "pipeline", 2)

	sql, args := ib.Build()
	if !strings.Contains(sql, "INSERT INTO workflows (id, name, version) VALUES ($1, $2, $3), ($4, $5, $6)") {
		t.Errorf("unexpected insert sql: %s", sql)
	}
	if len(args) != 6 {
		t.Fatalf("expected 6 args, got %d", len(args))
	}
}
