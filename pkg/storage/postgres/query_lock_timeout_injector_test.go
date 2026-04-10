package postgres

import (
	"context"
	"testing"
	"time"
)

func TestQueryLockTimeoutInjector(t *testing.T) {
	injector := NewQueryLockTimeoutInjector(1500 * time.Millisecond)

	if injector.DefaultTimeout() != 1500*time.Millisecond {
		t.Errorf("expected 1500ms default, got %v", injector.DefaultTimeout())
	}

	err := injector.InjectLockTimeout(context.Background(), nil, 2*time.Second)
	if err != nil {
		t.Fatalf("unexpected error with nil tx: %v", err)
	}
}
