package postgres

import (
	"context"
	"testing"
	"time"
)

func TestQueryStatementTimeoutManager(t *testing.T) {
	policy := DynamicTimeoutPolicy{
		InteractiveTimeout: 2 * time.Second,
		BatchETLTimeout:    30 * time.Second,
	}

	mgr := NewQueryStatementTimeoutManager(policy)
	if mgr.Policy().InteractiveTimeout != 2*time.Second {
		t.Errorf("expected 2s timeout, got %v", mgr.Policy().InteractiveTimeout)
	}

	// Safe application with nil tx
	err := mgr.ApplyTxTimeout(context.Background(), nil, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error with nil tx: %v", err)
	}
}
