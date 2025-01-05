package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type DrainPhase int32

const (
	DrainPhaseNormal DrainPhase = iota
	DrainPhaseStoppingIngress
	DrainPhaseAwaitingActive
	DrainPhaseFlushing
	DrainPhaseCompleted
)

func (p DrainPhase) String() string {
	switch p {
	case DrainPhaseNormal:
		return "NORMAL"
	case DrainPhaseStoppingIngress:
		return "STOPPING_INGRESS"
	case DrainPhaseAwaitingActive:
		return "AWAITING_ACTIVE"
	case DrainPhaseFlushing:
		return "FLUSHING"
	case DrainPhaseCompleted:
		return "COMPLETED"
	default:
		return "UNKNOWN"
	}
}

type DrainCoordinator struct {
	phase        int32
	activeTasks  int64
	doneCh       chan struct{}
	mu           sync.Mutex
	flushHooks   []func(context.Context) error
}

func NewDrainCoordinator() *DrainCoordinator {
	return &DrainCoordinator{
		phase:   int32(DrainPhaseNormal),
		doneCh:  make(chan struct{}),
	}
}

func (dc *DrainCoordinator) RegisterFlushHook(fn func(context.Context) error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.flushHooks = append(dc.flushHooks, fn)
}

func (dc *DrainCoordinator) IncActive() bool {
	if atomic.LoadInt32(&dc.phase) != int32(DrainPhaseNormal) {
		return false
	}
	atomic.AddInt64(&dc.activeTasks, 1)
	return true
}

func (dc *DrainCoordinator) DecActive() {
	atomic.AddInt64(&dc.activeTasks, -1)
}

func (dc *DrainCoordinator) Phase() DrainPhase {
	return DrainPhase(atomic.LoadInt32(&dc.phase))
}

func (dc *DrainCoordinator) StartDrain(ctx context.Context, timeout time.Duration) error {
	if !atomic.CompareAndSwapInt32(&dc.phase, int32(DrainPhaseNormal), int32(DrainPhaseStoppingIngress)) {
		return fmt.Errorf("drain already initiated")
	}

	deadline := time.Now().Add(timeout)
	atomic.StoreInt32(&dc.phase, int32(DrainPhaseAwaitingActive))

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		if atomic.LoadInt64(&dc.activeTasks) <= 0 {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("drain timeout waiting for active tasks")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}

	atomic.StoreInt32(&dc.phase, int32(DrainPhaseFlushing))
	dc.mu.Lock()
	hooks := append([]func(context.Context) error(nil), dc.flushHooks...)
	dc.mu.Unlock()

	for _, h := range hooks {
		if err := h(ctx); err != nil {
			return fmt.Errorf("flush hook failed: %w", err)
		}
	}

	atomic.StoreInt32(&dc.phase, int32(DrainPhaseCompleted))
	close(dc.doneCh)
	return nil
}

func (dc *DrainCoordinator) Done() <-chan struct{} {
	return dc.doneCh
}
