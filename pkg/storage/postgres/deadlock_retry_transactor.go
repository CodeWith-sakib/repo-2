package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

// DeadlockRetryTransactor wraps database transactions with exponential backoff on PostgreSQL 40P01 deadlocks.
type DeadlockRetryTransactor struct {
	db          *sql.DB
	maxRetries  int
	initialWait time.Duration
}

// NewDeadlockRetryTransactor creates a deadlock retry transactor.
func NewDeadlockRetryTransactor(db *sql.DB, maxRetries int, initialWait time.Duration) *DeadlockRetryTransactor {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	if initialWait <= 0 {
		initialWait = 20 * time.Millisecond
	}
	return &DeadlockRetryTransactor{
		db:          db,
		maxRetries:  maxRetries,
		initialWait: initialWait,
	}
}

// IsDeadlockError checks if error string indicates PostgreSQL deadlock or serialization failure.
func IsDeadlockError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "deadlock detected") ||
		strings.Contains(s, "40p01") ||
		strings.Contains(s, "could not serialize access") ||
		strings.Contains(s, "40001")
}

// ExecuteInTx executes fn inside a transaction, retrying up to maxRetries on deadlock.
func (d *DeadlockRetryTransactor) ExecuteInTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	if d.db == nil {
		// Mock pass-through for test environments
		return fn(nil)
	}

	wait := d.initialWait

	for attempt := 0; attempt <= d.maxRetries; attempt++ {
		tx, err := d.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		err = fn(tx)
		if err != nil {
			_ = tx.Rollback()
			if IsDeadlockError(err) && attempt < d.maxRetries {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(wait):
					wait *= 2
					continue
				}
			}
			return err
		}

		if err := tx.Commit(); err != nil {
			if IsDeadlockError(err) && attempt < d.maxRetries {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(wait):
					wait *= 2
					continue
				}
			}
			return err
		}

		return nil
	}

	return errors.New("exceeded maximum transaction retry attempts")
}
