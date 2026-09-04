package postgres

import (
	"testing"
)

func TestLockModeAnalyzer(t *testing.T) {
	analyzer := NewLockModeAnalyzer()

	l1 := analyzer.AnalyzeStatement("SELECT * FROM kf_workflows")
	if l1 != LockAccessShare {
		t.Errorf("expected ACCESS_SHARE, got %v", l1)
	}

	l2 := analyzer.AnalyzeStatement("UPDATE kf_step_runs SET status = 'DONE' WHERE id = '1'")
	if l2 != LockRowExclusive {
		t.Errorf("expected ROW_EXCLUSIVE, got %v", l2)
	}

	l3 := analyzer.AnalyzeStatement("ALTER TABLE kf_workflows ADD COLUMN flags jsonb")
	if l3 != LockAccessExclusive {
		t.Errorf("expected ACCESS_EXCLUSIVE, got %v", l3)
	}
	if !analyzer.IsExclusiveConflict(l3) {
		t.Error("ACCESS_EXCLUSIVE must conflict with reads")
	}
}
