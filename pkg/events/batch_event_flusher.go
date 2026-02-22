package events

import (
	"context"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

// BatchFlushHandler defines consumer sink for buffered events.
type BatchFlushHandler func(ctx context.Context, batch []*core.Event) error

// BatchEventFlusher buffers individual incoming events and flushes by count or time interval.
type BatchEventFlusher struct {
	mu           sync.Mutex
	buffer       []*core.Event
	maxBatchSize int
	flushInterval time.Duration
	handler      BatchFlushHandler
	stopCh       chan struct{}
}

// NewBatchEventFlusher constructs an asynchronous batch event flusher.
func NewBatchEventFlusher(maxBatchSize int, interval time.Duration, handler BatchFlushHandler) *BatchEventFlusher {
	if maxBatchSize <= 0 {
		maxBatchSize = 100
	}
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}
	return &BatchEventFlusher{
		buffer:        make([]*core.Event, 0, maxBatchSize),
		maxBatchSize:  maxBatchSize,
		flushInterval: interval,
		handler:       handler,
		stopCh:        make(chan struct{}),
	}
}

// Push adds an event to the buffer, triggering a flush if capacity is reached.
func (f *BatchEventFlusher) Push(ctx context.Context, evt *core.Event) error {
	f.mu.Lock()
	f.buffer = append(f.buffer, evt)
	shouldFlush := len(f.buffer) >= f.maxBatchSize
	f.mu.Unlock()

	if shouldFlush {
		return f.Flush(ctx)
	}
	return nil
}

// Flush immediately drains the active buffer into the sink handler.
func (f *BatchEventFlusher) Flush(ctx context.Context) error {
	f.mu.Lock()
	if len(f.buffer) == 0 {
		f.mu.Unlock()
		return nil
	}

	batch := f.buffer
	f.buffer = make([]*core.Event, 0, f.maxBatchSize)
	f.mu.Unlock()

	if f.handler != nil {
		return f.handler(ctx, batch)
	}
	return nil
}

// Start initiates background ticker to flush periodically.
func (f *BatchEventFlusher) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(f.flushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-f.stopCh:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = f.Flush(ctx)
			}
		}
	}()
}

// Stop terminates background flusher loop.
func (f *BatchEventFlusher) Stop() {
	f.mu.Lock()
	defer f.mu.Unlock()

	select {
	case <-f.stopCh:
	default:
		close(f.stopCh)
	}
}
