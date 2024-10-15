package retry

import (
	"context"
	"errors"
	"fmt"
	"time"

)

type NonRetryableError struct {
	Err error
}

func (e *NonRetryableError) Error() string {
	return fmt.Sprintf("non-retryable error: %v", e.Err)
}

func (e *NonRetryableError) Unwrap() error {
	return e.Err
}

func MarkNonRetryable(err error) error {
	if err == nil {
		return nil
	}
	return &NonRetryableError{Err: err}
}

func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	var nr *NonRetryableError
	return !errors.As(err, &nr)
}

type Policy struct {
	MaxAttempts int
	Backoff     BackoffStrategy
}

func NewPolicy(maxAttempts int, backoff BackoffStrategy) *Policy {
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	return &Policy{
		MaxAttempts: maxAttempts,
		Backoff:     backoff,
	}
}

func (p *Policy) Execute(ctx context.Context, op func(ctx context.Context, attempt int) error) error {
	var lastErr error

	for attempt := 1; attempt <= p.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := op(ctx, attempt)
		if err == nil {
			return nil
		}

		lastErr = err
		if !IsRetryable(err) || attempt >= p.MaxAttempts {
			return lastErr
		}

		sleepDuration := p.Backoff.NextInterval(attempt)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleepDuration):
		}
	}

	return lastErr
}
