package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

type panicExecutor struct{}

func (p *panicExecutor) Execute(ctx context.Context, sctx StepContext) (*StepResult, error) {
	panic("unexpected fatal failure")
}

type slowExecutor struct{}

func (s *slowExecutor) Execute(ctx context.Context, sctx StepContext) (*StepResult, error) {
	select {
	case <-time.After(100 * time.Millisecond):
		return &StepResult{Output: json.RawMessage(`{}`)}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestSandboxedTaskExecutor_Success(t *testing.T) {
	inner := &StaticJSONExecutor{Payload: []byte(`{"status":"ok"}`)}
	limits := DefaultSandboxLimits()
	sb := NewSandboxedTaskExecutor(inner, limits)

	ctx := context.Background()
	sctx := StepContext{RunID: "run-1", StepID: "step-1"}

	res, err := sb.Execute(ctx, sctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(res.Output) != `{"status":"ok"}` {
		t.Errorf("unexpected output: %s", string(res.Output))
	}

	total, succ, timeouts, panics, _ := sb.Metrics()
	if total != 1 || succ != 1 || timeouts != 0 || panics != 0 {
		t.Errorf("metrics mismatch: total=%d succ=%d timeouts=%d panics=%d", total, succ, timeouts, panics)
	}
}

func TestSandboxedTaskExecutor_PanicCatch(t *testing.T) {
	inner := &panicExecutor{}
	limits := DefaultSandboxLimits()
	sb := NewSandboxedTaskExecutor(inner, limits)

	ctx := context.Background()
	sctx := StepContext{RunID: "run-2", StepID: "step-2"}

	res, err := sb.Execute(ctx, sctx)
	if err == nil {
		t.Fatal("expected error from panic, got nil")
	}
	if res == nil || res.ErrorMessage == "" {
		t.Errorf("expected error message in result, got %+v", res)
	}

	_, _, _, panics, _ := sb.Metrics()
	if panics != 1 {
		t.Errorf("expected 1 panic recorded, got %d", panics)
	}
}

func TestSandboxedTaskExecutor_Timeout(t *testing.T) {
	inner := &slowExecutor{}
	limits := SandboxLimits{
		Timeout:        20 * time.Millisecond,
		MaxOutputBytes: 1024,
		CatchPanics:    true,
	}
	sb := NewSandboxedTaskExecutor(inner, limits)

	ctx := context.Background()
	sctx := StepContext{RunID: "run-3", StepID: "step-3"}

	res, err := sb.Execute(ctx, sctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
	if res == nil || !res.Retryable {
		t.Errorf("expected retryable result on timeout, got %+v", res)
	}

	_, _, timeouts, _, _ := sb.Metrics()
	if timeouts != 1 {
		t.Errorf("expected 1 timeout recorded, got %d", timeouts)
	}
}
