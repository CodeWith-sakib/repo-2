package retry

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestExponentialBackoffBounds(t *testing.T) {
	initial := 10 * time.Millisecond
	max := 100 * time.Millisecond
	eb := NewExponentialBackoff(initial, max, 2.0, false)

	i1 := eb.NextInterval(1)
	if i1 != initial {
		t.Errorf("attempt 1: got %v, expected %v", i1, initial)
	}

	i2 := eb.NextInterval(2)
	if i2 != 20*time.Millisecond {
		t.Errorf("attempt 2: got %v, expected %v", i2, 20*time.Millisecond)
	}

	iMax := eb.NextInterval(10)
	if iMax != max {
		t.Errorf("attempt 10: got %v, expected max %v", iMax, max)
	}
}

func TestExponentialBackoffJitterConcurrency(t *testing.T) {
	eb := NewExponentialBackoff(10*time.Millisecond, 200*time.Millisecond, 2.0, true)
	var wg sync.WaitGroup
	workers := 10

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for attempt := 1; attempt <= 20; attempt++ {
				interval := eb.NextInterval(attempt)
				if interval < 0 || interval > 200*time.Millisecond {
					t.Errorf("jitter interval out of bounds: %v", interval)
				}
			}
		}()
	}
	wg.Wait()
}

func TestPolicyRetrySuccess(t *testing.T) {
	eb := NewExponentialBackoff(time.Millisecond, 5*time.Millisecond, 2.0, false)
	policy := NewPolicy(3, eb)

	calls := 0
	err := policy.Execute(context.Background(), func(ctx context.Context, attempt int) error {
		calls++
		if attempt < 2 {
			return errors.New("transient error")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestPolicyNonRetryableError(t *testing.T) {
	policy := NewPolicy(5, NewExponentialBackoff(time.Millisecond, 5*time.Millisecond, 2.0, false))

	calls := 0
	fatalErr := errors.New("fatal configuration error")
	err := policy.Execute(context.Background(), func(ctx context.Context, attempt int) error {
		calls++
		return MarkNonRetryable(fatalErr)
	})

	if !errors.Is(err, fatalErr) {
		t.Fatalf("expected fatal error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected execution to stop after 1 attempt, got %d", calls)
	}
}
