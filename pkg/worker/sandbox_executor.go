package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// SandboxLimits defines execution boundaries for a sandboxed task.
type SandboxLimits struct {
	Timeout          time.Duration
	MaxOutputBytes   int
	CatchPanics      bool
	MaxRetryAttempts int
}

// DefaultSandboxLimits returns standard conservative execution limits.
func DefaultSandboxLimits() SandboxLimits {
	return SandboxLimits{
		Timeout:          30 * time.Second,
		MaxOutputBytes:   10 * 1024 * 1024, // 10 MB
		CatchPanics:      true,
		MaxRetryAttempts: 3,
	}
}

// SandboxMetrics tracks invocation telemetry.
type SandboxMetrics struct {
	TotalExecutions atomic.Int64
	Successes       atomic.Int64
	Timeouts        atomic.Int64
	Panics          atomic.Int64
	OutputExceeded  atomic.Int64
}

// ExecutionHook is called before and after task execution.
type ExecutionHook interface {
	BeforeExecute(ctx context.Context, sctx StepContext) context.Context
	AfterExecute(ctx context.Context, sctx StepContext, res *StepResult, err error, duration time.Duration)
}

// SandboxedTaskExecutor wraps any TaskExecutor with timeout, panic boundary, output limits, and hooks.
type SandboxedTaskExecutor struct {
	inner   TaskExecutor
	limits  SandboxLimits
	metrics SandboxMetrics
	mu      sync.RWMutex
	hooks   []ExecutionHook
}

// NewSandboxedTaskExecutor constructs a new sandboxed executor.
func NewSandboxedTaskExecutor(inner TaskExecutor, limits SandboxLimits) *SandboxedTaskExecutor {
	return &SandboxedTaskExecutor{
		inner:  inner,
		limits: limits,
	}
}

// AddHook registers an execution lifecycle hook.
func (e *SandboxedTaskExecutor) AddHook(hook ExecutionHook) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.hooks = append(e.hooks, hook)
}

// Execute enforces boundaries on the inner executor.
func (e *SandboxedTaskExecutor) Execute(ctx context.Context, sctx StepContext) (res *StepResult, err error) {
	e.metrics.TotalExecutions.Add(1)
	start := time.Now()

	e.mu.RLock()
	hooks := make([]ExecutionHook, len(e.hooks))
	copy(hooks, e.hooks)
	e.mu.RUnlock()

	for _, h := range hooks {
		ctx = h.BeforeExecute(ctx, sctx)
	}

	defer func() {
		dur := time.Since(start)

		if err == nil && res != nil && len(res.Output) > e.limits.MaxOutputBytes {
			e.metrics.OutputExceeded.Add(1)
			err = fmt.Errorf("task output (%d bytes) exceeded limit of %d bytes", len(res.Output), e.limits.MaxOutputBytes)
			res = &StepResult{
				ErrorMessage: err.Error(),
				Retryable:    false,
			}
		}

		if err == nil && (res == nil || res.ErrorMessage == "") {
			e.metrics.Successes.Add(1)
		}

		for _, h := range hooks {
			h.AfterExecute(ctx, sctx, res, err, dur)
		}
	}()

	execCtx := ctx
	var cancel context.CancelFunc
	if e.limits.Timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, e.limits.Timeout)
		defer cancel()
	}

	resCh := make(chan *StepResult, 1)
	errCh := make(chan error, 1)

	go func() {
		defer func() {
			if e.limits.CatchPanics {
				if r := recover(); r != nil {
					e.metrics.Panics.Add(1)
					panicErr := fmt.Errorf("task panicked: %v", r)
					resCh <- &StepResult{
						ErrorMessage: panicErr.Error(),
						Retryable:    false,
					}
					errCh <- panicErr
				}
			}
		}()

		r, execErr := e.inner.Execute(execCtx, sctx)
		resCh <- r
		errCh <- execErr
	}()

	select {
	case <-execCtx.Done():
		if execCtx.Err() == context.DeadlineExceeded {
			e.metrics.Timeouts.Add(1)
			return &StepResult{
				ErrorMessage: fmt.Sprintf("task execution timed out after %s", e.limits.Timeout),
				Retryable:    true,
			}, context.DeadlineExceeded
		}
		return nil, execCtx.Err()

	case err = <-errCh:
		res = <-resCh
		return res, err
	}
}

// Metrics returns telemetry counters.
func (e *SandboxedTaskExecutor) Metrics() (total, success, timeouts, panics, outputExceeded int64) {
	return e.metrics.TotalExecutions.Load(),
		e.metrics.Successes.Load(),
		e.metrics.Timeouts.Load(),
		e.metrics.Panics.Load(),
		e.metrics.OutputExceeded.Load()
}

// StaticJSONExecutor is a helper task executor returning static JSON payload.
type StaticJSONExecutor struct {
	Payload []byte
}

func (s *StaticJSONExecutor) Execute(ctx context.Context, sctx StepContext) (*StepResult, error) {
	return &StepResult{
		Output: json.RawMessage(s.Payload),
	}, nil
}
