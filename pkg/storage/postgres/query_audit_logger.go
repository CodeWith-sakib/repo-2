package postgres

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

// QueryAuditRecord captures SQL audit trail for compliance.
type QueryAuditRecord struct {
	QueryID      string        `json:"query_id"`
	SQL          string        `json:"sql"`
	Duration     time.Duration `json:"duration"`
	RowsAffected int64         `json:"rows_affected"`
	Error        string        `json:"error,omitempty"`
	ExecutedAt   time.Time     `json:"executed_at"`
}

// QueryAuditLogger records execution trails in memory or forwards to storage.
type QueryAuditLogger struct {
	mu      sync.Mutex
	records []QueryAuditRecord
	maxLogs int
	db      *sql.DB
}

// NewQueryAuditLogger creates an audit logger with maximum in-memory ring buffer.
func NewQueryAuditLogger(db *sql.DB, maxLogs int) *QueryAuditLogger {
	if maxLogs <= 0 {
		maxLogs = 1000
	}
	return &QueryAuditLogger{
		records: make([]QueryAuditRecord, 0, maxLogs),
		maxLogs: maxLogs,
		db:      db,
	}
}

// Record captures a single executed query entry.
func (l *QueryAuditLogger) Record(ctx context.Context, rec QueryAuditRecord) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.records) >= l.maxLogs {
		// Evict oldest record
		l.records = l.records[1:]
	}
	l.records = append(l.records, rec)
}

// RecentRecords returns a slice copy of recently captured audit records.
func (l *QueryAuditLogger) RecentRecords(limit int) []QueryAuditRecord {
	l.mu.Lock()
	defer l.mu.Unlock()

	n := len(l.records)
	if limit <= 0 || limit > n {
		limit = n
	}

	result := make([]QueryAuditRecord, limit)
	copy(result, l.records[n-limit:])
	return result
}

// Count returns total currently buffered audit records.
func (l *QueryAuditLogger) Count() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.records)
}
