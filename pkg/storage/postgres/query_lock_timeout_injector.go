package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// QueryLockTimeoutInjector sets transaction-level lock_timeout to fail fast instead of queueing behind locks.
type QueryLockTimeoutInjector struct {
	defaultLockTimeout time.Duration
}

// NewQueryLockTimeoutInjector creates a lock timeout injector.
func NewQueryLockTimeoutInjector(timeout time.Duration) *QueryLockTimeoutInjector {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &QueryLockTimeoutInjector{
		defaultLockTimeout: timeout,
	}
}

// InjectLockTimeout executes SET LOCAL lock_timeout on a transaction.
func (i *QueryLockTimeoutInjector) InjectLockTimeout(ctx context.Context, tx *sql.Tx, customTimeout time.Duration) error {
	if tx == nil {
		return nil
	}

	dur := i.defaultLockTimeout
	if customTimeout > 0 {
		dur = customTimeout
	}

	query := fmt.Sprintf("SET LOCAL lock_timeout = '%dms';", dur.Milliseconds())
	_, err := tx.ExecContext(ctx, query)
	return err
}

// DefaultTimeout returns configured baseline lock timeout.
func (i *QueryLockTimeoutInjector) DefaultTimeout() time.Duration {
	return i.defaultLockTimeout
}
