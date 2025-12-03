package postgres

import (
	"fmt"
	"hash/fnv"
)

// AdvisoryLockScope defines whether a lock is held for the session or transaction lifetime.
type AdvisoryLockScope string

const (
	LockScopeSession     AdvisoryLockScope = "SESSION"
	LockScopeTransaction AdvisoryLockScope = "TRANSACTION"
)

// AdvisoryLockHelper generates PostgreSQL advisory lock acquisition and release queries.
type AdvisoryLockHelper struct{}

// NewAdvisoryLockHelper creates an advisory lock helper.
func NewAdvisoryLockHelper() *AdvisoryLockHelper {
	return &AdvisoryLockHelper{}
}

// HashResourceKey hashes an arbitrary string identifier into a signed 64-bit integer.
func (h *AdvisoryLockHelper) HashResourceKey(resource string) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(resource))
	return int64(hasher.Sum64())
}

// LockQuery returns SQL to acquire an advisory lock (blocking).
func (h *AdvisoryLockHelper) LockQuery(key int64, scope AdvisoryLockScope) string {
	if scope == LockScopeTransaction {
		return fmt.Sprintf("SELECT pg_advisory_xact_lock(%d);", key)
	}
	return fmt.Sprintf("SELECT pg_advisory_lock(%d);", key)
}

// TryLockQuery returns SQL to attempt acquiring an advisory lock non-blockingly (returns boolean).
func (h *AdvisoryLockHelper) TryLockQuery(key int64, scope AdvisoryLockScope) string {
	if scope == LockScopeTransaction {
		return fmt.Sprintf("SELECT pg_try_advisory_xact_lock(%d);", key)
	}
	return fmt.Sprintf("SELECT pg_try_advisory_lock(%d);", key)
}

// UnlockQuery returns SQL to explicitly release a session-scoped advisory lock.
func (h *AdvisoryLockHelper) UnlockQuery(key int64) string {
	return fmt.Sprintf("SELECT pg_advisory_unlock(%d);", key)
}

// RowLockForUpdate returns SELECT ... FOR UPDATE [NOWAIT | SKIP LOCKED].
func (h *AdvisoryLockHelper) RowLockForUpdate(table string, idCol string, idVal interface{}, noWait, skipLocked bool) (string, []interface{}) {
	clause := "FOR UPDATE"
	if noWait {
		clause = "FOR UPDATE NOWAIT"
	} else if skipLocked {
		clause = "FOR UPDATE SKIP LOCKED"
	}
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = $1 %s;", table, idCol, clause)
	return query, []interface{}{idVal}
}
