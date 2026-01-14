package worker

import (
	"context"
	"sync"
	"time"
)

// GracefulTerminator coordinates orderly worker shutdown and task draining.
type GracefulTerminator struct {
	mu           sync.Mutex
	activeCount  int
	shuttingDown bool
	drainNotify  chan struct{}
}

// NewGracefulTerminator creates a graceful terminator coordinator.
func NewGracefulTerminator() *GracefulTerminator {
	return &GracefulTerminator{
		drainNotify: make(chan struct{}, 1),
	}
}

// TryAcquire attempts to register an active task before shutdown begins.
func (gt *GracefulTerminator) TryAcquire() bool {
	gt.mu.Lock()
	defer gt.mu.Unlock()

	if gt.shuttingDown {
		return false
	}
	gt.activeCount++
	return true
}

// Release decrements active task count and signals drain if zero during shutdown.
func (gt *GracefulTerminator) Release() {
	gt.mu.Lock()
	defer gt.mu.Unlock()

	if gt.activeCount > 0 {
		gt.activeCount--
	}
	if gt.shuttingDown && gt.activeCount == 0 {
		select {
		case gt.drainNotify <- struct{}{}:
		default:
		}
	}
}

// Shutdown initiates shutdown mode and waits up to timeout for active tasks to reach 0.
func (gt *GracefulTerminator) Shutdown(ctx context.Context, timeout time.Duration) error {
	gt.mu.Lock()
	gt.shuttingDown = true
	if gt.activeCount == 0 {
		gt.mu.Unlock()
		return nil
	}
	gt.mu.Unlock()

	ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctxTimeout.Done():
			return ctxTimeout.Err()
		case <-gt.drainNotify:
			gt.mu.Lock()
			count := gt.activeCount
			gt.mu.Unlock()
			if count == 0 {
				return nil
			}
		}
	}
}

// IsShuttingDown checks if shutdown has been initiated.
func (gt *GracefulTerminator) IsShuttingDown() bool {
	gt.mu.Lock()
	defer gt.mu.Unlock()
	return gt.shuttingDown
}

// ActiveCount returns current running tasks count.
func (gt *GracefulTerminator) ActiveCount() int {
	gt.mu.Lock()
	defer gt.mu.Unlock()
	return gt.activeCount
}
