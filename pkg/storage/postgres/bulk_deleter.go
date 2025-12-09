package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// BulkDeleteOptions controls safety and batch sizing during massive row pruning.
type BulkDeleteOptions struct {
	TableName    string
	WhereClause  string
	BatchSize    int
	SleepBetween time.Duration
	DryRun       bool
}

// BulkDeleteResult reports statistics on chunked row deletions.
type BulkDeleteResult struct {
	TotalDeleted int64
	BatchesRun   int
	Duration     time.Duration
}

// BulkDeleter executes batched deletes in chunks to prevent WAL bloat and excessive lock contention.
type BulkDeleter struct {
	db *sql.DB
}

// NewBulkDeleter creates a new chunked bulk deleter.
func NewBulkDeleter(db *sql.DB) *BulkDeleter {
	return &BulkDeleter{db: db}
}

// DeleteInBatches executes chunked deletions using a subquery pattern (e.g., ctid IN (SELECT ctid FROM ... LIMIT N)).
func (bd *BulkDeleter) DeleteInBatches(ctx context.Context, opts BulkDeleteOptions) (*BulkDeleteResult, error) {
	if opts.TableName == "" {
		return nil, fmt.Errorf("table name is required")
	}
	if opts.BatchSize <= 0 {
		opts.BatchSize = 500
	}
	if opts.WhereClause == "" {
		return nil, fmt.Errorf("where clause is required to prevent accidental full table wipe")
	}

	start := time.Now()
	res := &BulkDeleteResult{}

	// If database is nil (e.g. testing or simulated execution mode), perform dry-run logic
	if bd.db == nil || opts.DryRun {
		res.Duration = time.Since(start)
		res.TotalDeleted = 0
		res.BatchesRun = 0
		return res, nil
	}

	deleteSQL := fmt.Sprintf(
		"DELETE FROM %s WHERE ctid IN (SELECT ctid FROM %s WHERE %s LIMIT %d)",
		opts.TableName, opts.TableName, opts.WhereClause, opts.BatchSize,
	)

	for {
		select {
		case <-ctx.Done():
			return res, ctx.Err()
		default:
		}

		sqlRes, err := bd.db.ExecContext(ctx, deleteSQL)
		if err != nil {
			return res, fmt.Errorf("failed executing batch delete: %w", err)
		}

		rowsAffected, err := sqlRes.RowsAffected()
		if err != nil {
			return res, fmt.Errorf("failed retrieving affected rows count: %w", err)
		}

		res.TotalDeleted += rowsAffected
		res.BatchesRun++

		if rowsAffected < int64(opts.BatchSize) {
			break // No more matching rows
		}

		if opts.SleepBetween > 0 {
			time.Sleep(opts.SleepBetween)
		}
	}

	res.Duration = time.Since(start)
	return res, nil
}
