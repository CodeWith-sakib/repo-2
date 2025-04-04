package postgres

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDBCircuitBreaker(t *testing.T) {
	cb := NewDBCircuitBreaker(2, 50*time.Millisecond)
	failOp := func() error { return errors.New("connection refused") }
	passOp := func() error { return nil }

	ctx := context.Background()

	// 1st fail
	_ = cb.Execute(ctx, failOp)
	if cb.State() != StateClosed {
		t.Errorf("expected closed after 1 fail, got %v", cb.State())
	}

	// 2nd fail trips to open
	_ = cb.Execute(ctx, failOp)
	if cb.State() != StateOpen {
		t.Errorf("expected open after 2 fails, got %v", cb.State())
	}

	// While open, should fail immediately with ErrCircuitOpen
	err := cb.Execute(ctx, passOp)
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}

	// Wait for reset timeout
	time.Sleep(60 * time.Millisecond)

	// Half-open attempt succeeds, restoring closed
	err = cb.Execute(ctx, passOp)
	if err != nil || cb.State() != StateClosed {
		t.Errorf("expected closed after recovery, got %v state and err: %v", cb.State(), err)
	}
}
