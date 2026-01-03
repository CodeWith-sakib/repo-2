package worker

import (
	"context"
	"sync"
	"time"
)

// StepProgressReport records live execution progress for long-running worker tasks.
type StepProgressReport struct {
	RunID         string                 `json:"run_id"`
	StepID        string                 `json:"step_id"`
	ProgressPct   float64                `json:"progress_pct"`
	ProcessedRows int64                  `json:"processed_rows"`
	TotalRows     int64                  `json:"total_rows"`
	StatusMessage string                 `json:"status_message"`
	LastUpdate    time.Time              `json:"last_update"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ProgressSink defines external reporting hook for heartbeats (e.g., storage or event bus).
type ProgressSink func(ctx context.Context, report StepProgressReport) error

// StepHeartbeatRecorder manages recurring background heartbeats and progress emission.
type StepHeartbeatRecorder struct {
	mu       sync.Mutex
	report   StepProgressReport
	sink     ProgressSink
	interval time.Duration
	stopCh   chan struct{}
}

// NewStepHeartbeatRecorder initializes a periodic heartbeat recorder.
func NewStepHeartbeatRecorder(runID, stepID string, interval time.Duration, sink ProgressSink) *StepHeartbeatRecorder {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &StepHeartbeatRecorder{
		report: StepProgressReport{
			RunID:      runID,
			StepID:     stepID,
			LastUpdate: time.Now().UTC(),
			Metadata:   make(map[string]interface{}),
		},
		sink:     sink,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// UpdateProgress updates internal progress metrics thread-safely.
func (r *StepHeartbeatRecorder) UpdateProgress(pct float64, processed, total int64, msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.report.ProgressPct = pct
	r.report.ProcessedRows = processed
	r.report.TotalRows = total
	r.report.StatusMessage = msg
	r.report.LastUpdate = time.Now().UTC()
}

// Start begins periodic background heartbeat flushing.
func (r *StepHeartbeatRecorder) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()

		for {
			select {
			case <-r.stopCh:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.Flush(ctx)
			}
		}
	}()
}

// Flush executes an immediate sync flush to the progress sink.
func (r *StepHeartbeatRecorder) Flush(ctx context.Context) {
	r.mu.Lock()
	snapshot := r.report
	sink := r.sink
	r.mu.Unlock()

	if sink != nil {
		_ = sink(ctx, snapshot)
	}
}

// Stop terminates the background ticker.
func (r *StepHeartbeatRecorder) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	select {
	case <-r.stopCh:
	default:
		close(r.stopCh)
	}
}
