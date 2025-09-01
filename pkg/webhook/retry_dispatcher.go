package webhook

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// DispatchAttempt records the result of a single webhook delivery attempt.
type DispatchAttempt struct {
	AttemptNumber int
	SentAt        time.Time
	StatusCode    int
	Latency       time.Duration
	ErrorMsg      string
}

// WebhookEndpointConfig specifies delivery and retry parameters for a webhook target.
type WebhookEndpointConfig struct {
	URL            string
	MaxAttempts    int
	InitialBackoff time.Duration
	BackoffFactor  float64
	MaxBackoff     time.Duration
	Timeout        time.Duration
}

// RetryDispatchResult is the final outcome of a webhook delivery with attempt history.
type RetryDispatchResult struct {
	EndpointURL string
	Delivered   bool
	Attempts    []DispatchAttempt
	FinalError  string
}

// DeliveryFn is a function that sends a payload to a URL and returns HTTP status code.
type DeliveryFn func(ctx context.Context, url string, payload []byte) (statusCode int, err error)

// RetryDispatcher delivers webhook payloads with exponential backoff and jitter.
type RetryDispatcher struct {
	mu         sync.Mutex
	deliveryFn DeliveryFn
}

// NewRetryDispatcher creates a dispatcher using the given delivery function.
func NewRetryDispatcher(deliveryFn DeliveryFn) *RetryDispatcher {
	return &RetryDispatcher{deliveryFn: deliveryFn}
}

// Dispatch attempts to deliver payload to the endpoint with retries.
func (d *RetryDispatcher) Dispatch(ctx context.Context, cfg WebhookEndpointConfig, payload []byte) *RetryDispatchResult {
	result := &RetryDispatchResult{
		EndpointURL: cfg.URL,
	}

	backoff := cfg.InitialBackoff

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		start := time.Now()

		statusCode, err := d.deliveryFn(attemptCtx, cfg.URL, payload)
		cancel()

		latency := time.Since(start)
		rec := DispatchAttempt{
			AttemptNumber: attempt,
			SentAt:        start,
			StatusCode:    statusCode,
			Latency:       latency,
		}

		if err != nil {
			rec.ErrorMsg = err.Error()
		}

		result.Attempts = append(result.Attempts, rec)

		if err == nil && statusCode >= 200 && statusCode < 300 {
			result.Delivered = true
			return result
		}

		// Last attempt — don't wait
		if attempt == cfg.MaxAttempts {
			break
		}

		// Exponential backoff with cap
		waitDur := backoff
		if waitDur > cfg.MaxBackoff {
			waitDur = cfg.MaxBackoff
		}

		select {
		case <-ctx.Done():
			result.FinalError = "context cancelled during retry wait"
			return result
		case <-time.After(waitDur):
		}

		nextBackoff := time.Duration(float64(backoff) * cfg.BackoffFactor)
		backoff = time.Duration(math.Min(float64(nextBackoff), float64(cfg.MaxBackoff)))
	}

	if !result.Delivered {
		result.FinalError = fmt.Sprintf("all %d delivery attempts failed", cfg.MaxAttempts)
	}

	return result
}
