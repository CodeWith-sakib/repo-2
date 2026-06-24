package postgres

import (
	"fmt"
	"sync"
	"time"
)

// AdaptiveTimeoutConfig governs query timeouts according to priority and query complexity.
type AdaptiveTimeoutConfig struct {
	InteractiveTimeout time.Duration
	BatchTimeout       time.Duration
	AnalyticalTimeout  time.Duration
}

// StatementTimeoutAdapter configures dynamic SET statement_timeout according to execution intent.
type StatementTimeoutAdapter struct {
	mu     sync.RWMutex
	config AdaptiveTimeoutConfig
}

// NewStatementTimeoutAdapter creates an adapter.
func NewStatementTimeoutAdapter(cfg AdaptiveTimeoutConfig) *StatementTimeoutAdapter {
	if cfg.InteractiveTimeout <= 0 {
		cfg.InteractiveTimeout = 3 * time.Second
	}
	if cfg.BatchTimeout <= 0 {
		cfg.BatchTimeout = 30 * time.Second
	}
	if cfg.AnalyticalTimeout <= 0 {
		cfg.AnalyticalTimeout = 5 * time.Minute
	}
	return &StatementTimeoutAdapter{
		config: cfg,
	}
}

// TimeoutForPriority returns recommended statement timeout in milliseconds.
func (a *StatementTimeoutAdapter) TimeoutForPriority(priority string) time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()

	switch priority {
	case "INTERACTIVE", "REALTIME":
		return a.config.InteractiveTimeout
	case "BATCH", "BACKGROUND":
		return a.config.BatchTimeout
	case "ANALYTICS", "REPORTING":
		return a.config.AnalyticalTimeout
	default:
		return a.config.BatchTimeout
	}
}

// BuildSetStatementSQL produces SQL command to set statement_timeout for current connection.
func (a *StatementTimeoutAdapter) BuildSetStatementSQL(priority string) string {
	dur := a.TimeoutForPriority(priority)
	ms := int64(dur / time.Millisecond)
	return fmt.Sprintf("SET statement_timeout = %d;", ms)
}
