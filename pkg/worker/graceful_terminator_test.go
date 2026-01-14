package worker

import (
	"context"
	"testing"
	"time"
)

func TestGracefulTerminator(t *testing.T) {
	gt := NewGracefulTerminator()

	if !gt.TryAcquire() {
		t.Fatal("expected acquire to succeed")
	}
	if gt.ActiveCount() != 1 {
		t.Errorf("expected 1 active task, got %d", gt.ActiveCount())
	}

	// Release in background after 50ms
	go func() {
		time.Sleep(50 * time.Millisecond)
		gt.Release()
	}()

	err := gt.Shutdown(context.Background(), 500*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected shutdown error: %v", err)
	}

	if gt.ActiveCount() != 0 {
		t.Errorf("expected 0 active tasks, got %d", gt.ActiveCount())
	}

	// Subsequent acquire should be rejected
	if gt.TryAcquire() {
		t.Error("expected acquire after shutdown to fail")
	}
}
