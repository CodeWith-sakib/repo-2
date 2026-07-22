package postgres

import (
	"testing"
)

func TestPreparedStatementCache(t *testing.T) {
	cache := NewPreparedStatementCache(100)

	sql := "SELECT id, status FROM kf_workflows WHERE tenant_id = $1"
	name1, hit1 := cache.GetOrPrepare(sql)
	if hit1 {
		t.Error("expected cache miss on first encounter")
	}

	name2, hit2 := cache.GetOrPrepare(sql)
	if !hit2 {
		t.Error("expected cache hit on second encounter")
	}
	if name1 != name2 {
		t.Errorf("statement name mismatch: %s vs %s", name1, name2)
	}
	if cache.StatementCount() != 1 {
		t.Errorf("expected 1 statement cached, got %d", cache.StatementCount())
	}
}
