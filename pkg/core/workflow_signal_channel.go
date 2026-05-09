package core

import (
	"context"
	"errors"
	"sync"
	"time"
)

// WorkflowSignal represents an asynchronous external payload event delivered directly into a running workflow execution.
type WorkflowSignal struct {
	Name      string      `json:"name"`
	RunID     string      `json:"run_id"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// WorkflowSignalChannel routes external signals to suspended workflow step executions.
type WorkflowSignalChannel struct {
	mu       sync.RWMutex
	channels map[string]chan WorkflowSignal // runID:signalName -> chan
}

// NewWorkflowSignalChannel creates a signal delivery hub.
func NewWorkflowSignalChannel() *WorkflowSignalChannel {
	return &WorkflowSignalChannel{
		channels: make(map[string]chan WorkflowSignal),
	}
}

// Subscribe registers a receiving channel for a specific signal name on a workflow run.
func (sc *WorkflowSignalChannel) Subscribe(runID, signalName string, bufferSize int) <-chan WorkflowSignal {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if bufferSize <= 0 {
		bufferSize = 10
	}
	key := runID + ":" + signalName
	ch, exists := sc.channels[key]
	if !exists {
		ch = make(chan WorkflowSignal, bufferSize)
		sc.channels[key] = ch
	}
	return ch
}

// Send delivers a signal to waiting subscribers.
func (sc *WorkflowSignalChannel) Send(signal WorkflowSignal) error {
	if signal.RunID == "" || signal.Name == "" {
		return errors.New("signal RunID and Name are required")
	}

	sc.mu.RLock()
	key := signal.RunID + ":" + signal.Name
	ch, exists := sc.channels[key]
	sc.mu.RUnlock()

	if !exists {
		return errors.New("no active subscriber listening for signal")
	}

	select {
	case ch <- signal:
		return nil
	default:
		return errors.New("signal channel buffer is full")
	}
}

// WaitSignal blocks until target signal arrives or context deadline expires.
func (sc *WorkflowSignalChannel) WaitSignal(ctx context.Context, runID, signalName string) (WorkflowSignal, error) {
	ch := sc.Subscribe(runID, signalName, 1)

	select {
	case <-ctx.Done():
		return WorkflowSignal{}, ctx.Err()
	case sig := <-ch:
		return sig, nil
	}
}
