package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// DynamicTimeoutPolicy sets adaptive statement timeouts based on query workload type.
type DynamicTimeoutPolicy struct {
	InteractiveTimeout time.Duration
	BatchETLTimeout    time.Duration
	MigrationTimeout   time.Duration
}

// QueryStatementTimeoutManager injects safe session-level SET statement_timeout prior to query execution.
type QueryStatementTimeoutManager struct {
	policy DynamicTimeoutPolicy
}

// NewQueryStatementTimeoutManager creates a statement timeout manager.
func NewQueryStatementTimeoutManager(policy DynamicTimeoutPolicy) *QueryStatementTimeoutManager {
	if policy.InteractiveTimeout <= 0 {
		policy.InteractiveTimeout = 5 * time.Second
	}
	if policy.BatchETLTimeout <= 0 {
		policy.BatchETLTimeout = 5 * time.Minute
	}
	if policy.MigrationTimeout <= 0 {
		policy.MigrationTimeout = 15 * time.Minute
	}
	return &QueryStatementTimeoutManager{
		policy: policy,
	}
}

// ApplyTxTimeout sets local transaction-scoped statement timeout (SET LOCAL statement_timeout).
func (m *QueryStatementTimeoutManager) ApplyTxTimeout(ctx context.Context, tx *sql.Tx, timeout time.Duration) error {
	if tx == nil {
		return nil
	}
	if timeout <= 0 {
		timeout = m.policy.InteractiveTimeout
	}

	millis := timeout.Milliseconds()
	query := fmt.Sprintf("SET LOCAL statement_timeout = %d;", millis)
	_, err := tx.ExecContext(ctx, query)
	return err
}

// Policy returns current timeout policy configuration.
func (m *QueryStatementTimeoutManager) Policy() DynamicTimeoutPolicy {
	return m.policy
}
