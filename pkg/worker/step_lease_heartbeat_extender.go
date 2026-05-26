package worker

import (
	"context"
	"sync"
	"time"
)

// LeaseRenewHook defines callback to renew worker lease in persistent store.
type LeaseRenewHook func(ctx context.Context, stepRunID string, duration time.Duration) error

// StepLeaseHeartbeatExtender periodically renews step lease locks during active long-running executions.
type StepLeaseHeartbeatExtender struct {
	mu           sync.Mutex
	stepRunID    string
	renewHook    LeaseRenewHook
	interval     time.Duration
	leaseTTL     time.Duration
	stopCh       chan struct{}
	lastRenewed  time.Time
}

// NewStepLeaseHeartbeatExtender creates an automatic lease heartbeat extender.
func NewStepLeaseHeartbeatExtender(stepRunID string, interval, leaseTTL time.Duration, hook LeaseRenewHook) *StepLeaseHeartbeatExtender {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	if leaseTTL <= 0 {
		leaseTTL = 15 * time.Second
	}
	return &StepLeaseHeartbeatExtender{
		stepRunID: stepRunID,
		renewHook: hook,
		interval:  interval,
		leaseTTL:  leaseTTL,
		stopCh:    make(chan struct{}),
	}
}

// Start begins periodic background lease renewal.
func (e *StepLeaseHeartbeatExtender) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(e.interval)
		defer ticker.Stop()

		for {
			select {
			case <-e.stopCh:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				e.Renew(ctx)
			}
		}
	}()
}

// Renew executes an immediate renewal attempt.
func (e *StepLeaseHeartbeatExtender) Renew(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.renewHook != nil {
		if err := e.renewHook(ctx, e.stepRunID, e.leaseTTL); err != nil {
			return err
		}
	}
	e.lastRenewed = time.Now().UTC()
	return nil
}

// Stop terminates the renewal loop.
func (e *StepLeaseHeartbeatExtender) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	select {
	case <-e.stopCh:
	default:
		close(e.stopCh)
	}
}
