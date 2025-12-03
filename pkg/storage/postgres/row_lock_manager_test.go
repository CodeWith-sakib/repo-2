package postgres

import (
	"strings"
	"testing"
)

func TestAdvisoryLockHelper_HashAndQueries(t *testing.T) {
	h := NewAdvisoryLockHelper()

	res := "workflow:wf-payroll-2026"
	key := h.HashResourceKey(res)

	if key == 0 {
		t.Error("expected non-zero hash key")
	}

	// Session lock
	lockSQL := h.LockQuery(key, LockScopeSession)
	if !strings.Contains(lockSQL, "pg_advisory_lock") {
		t.Errorf("expected pg_advisory_lock in %s", lockSQL)
	}

	// Tx try-lock
	trySQL := h.TryLockQuery(key, LockScopeTransaction)
	if !strings.Contains(trySQL, "pg_try_advisory_xact_lock") {
		t.Errorf("expected pg_try_advisory_xact_lock in %s", trySQL)
	}

	// Unlock
	unlockSQL := h.UnlockQuery(key)
	if !strings.Contains(unlockSQL, "pg_advisory_unlock") {
		t.Errorf("expected pg_advisory_unlock in %s", unlockSQL)
	}

	// Row lock with SKIP LOCKED
	rowLockSQL, args := h.RowLockForUpdate("orders", "id", "ord-1", false, true)
	if !strings.Contains(rowLockSQL, "FOR UPDATE SKIP LOCKED") {
		t.Errorf("expected FOR UPDATE SKIP LOCKED in %s", rowLockSQL)
	}
	if len(args) != 1 || args[0] != "ord-1" {
		t.Errorf("unexpected args: %v", args)
	}
}
