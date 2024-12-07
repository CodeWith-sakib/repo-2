package retry

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreakerTripping(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		MaxFailures: 2,
		Timeout:     50 * time.Millisecond,
	})

	dummyErr := errors.New("remote service failure")

	// 1st failure -> opens circuit
	_ = cb.Execute(func() error { return dummyErr })

	// Immediate next call should trip
	err := cb.Execute(func() error { return nil })
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}

	// Sleep until timeout expires
	time.Sleep(60 * time.Millisecond)

	// Half-open attempt success
	_ = cb.Execute(func() error { return nil })
	// Second success closes circuit
	_ = cb.Execute(func() error { return nil })

	if cb.State() != StateClosed {
		t.Errorf("expected circuit state Closed, got %s", cb.State())
	}
}
