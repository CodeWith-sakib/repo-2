package postgres

import (
	"strings"
	"sync"
)

// PostgresLockLevel classifies SQL lock severity and concurrency impact.
type PostgresLockLevel string

const (
	LockAccessShare          PostgresLockLevel = "ACCESS_SHARE"           // SELECT
	LockRowShare             PostgresLockLevel = "ROW_SHARE"              // SELECT FOR UPDATE
	LockRowExclusive         PostgresLockLevel = "ROW_EXCLUSIVE"          // INSERT, UPDATE, DELETE
	LockShareUpdateExclusive PostgresLockLevel = "SHARE_UPDATE_EXCLUSIVE" // VACUUM, ANALYZE, CREATE INDEX CONCURRENTLY
	LockAccessExclusive      PostgresLockLevel = "ACCESS_EXCLUSIVE"       // ALTER TABLE, DROP TABLE, TRUNCATE
)

// LockModeAnalyzer inspects SQL statements to determine minimum lock tier required.
type LockModeAnalyzer struct {
	mu sync.RWMutex
}

// NewLockModeAnalyzer creates an analyzer.
func NewLockModeAnalyzer() *LockModeAnalyzer {
	return &LockModeAnalyzer{}
}

// AnalyzeStatement returns the estimated Postgres table lock level.
func (a *LockModeAnalyzer) AnalyzeStatement(sql string) PostgresLockLevel {
	a.mu.RLock()
	defer a.mu.RUnlock()

	upper := strings.ToUpper(strings.TrimSpace(sql))

	if strings.HasPrefix(upper, "ALTER TABLE") || strings.HasPrefix(upper, "DROP TABLE") || strings.HasPrefix(upper, "TRUNCATE") {
		return LockAccessExclusive
	}
	if strings.Contains(upper, "CREATE INDEX CONCURRENTLY") || strings.HasPrefix(upper, "VACUUM") || strings.HasPrefix(upper, "ANALYZE") {
		return LockShareUpdateExclusive
	}
	if strings.HasPrefix(upper, "INSERT") || strings.HasPrefix(upper, "UPDATE") || strings.HasPrefix(upper, "DELETE") {
		return LockRowExclusive
	}
	if strings.Contains(upper, "FOR UPDATE") || strings.Contains(upper, "FOR SHARE") {
		return LockRowShare
	}
	return LockAccessShare
}

// IsExclusiveConflict checks if mode blocks normal read traffic.
func (a *LockModeAnalyzer) IsExclusiveConflict(mode PostgresLockLevel) bool {
	return mode == LockAccessExclusive
}
