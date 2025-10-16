package core

import (
	"encoding/json"
	"fmt"
	"sync"
)

// FanOutPolicy defines error handling semantics for dynamic parallel step executions.
type FanOutPolicy string

const (
	PolicyFailFast       FanOutPolicy = "FAIL_FAST"
	PolicyCollectAll     FanOutPolicy = "COLLECT_ALL"
	PolicyIgnoreFailures FanOutPolicy = "IGNORE_FAILURES"
)

// FanOutConfig configures scatter-gather execution parameters.
type FanOutConfig struct {
	MaxParallelism int          `json:"max_parallelism"`
	Policy         FanOutPolicy `json:"policy"`
}

// DefaultFanOutConfig returns standard configuration with bounded concurrency.
func DefaultFanOutConfig() FanOutConfig {
	return FanOutConfig{
		MaxParallelism: 16,
		Policy:         PolicyCollectAll,
	}
}

// FanOutItemResult records the execution outcome of an individual scattered slice.
type FanOutItemResult struct {
	Index   int             `json:"index"`
	Item    interface{}     `json:"item"`
	Output  json.RawMessage `json:"output,omitempty"`
	Error   string          `json:"error,omitempty"`
	Success bool            `json:"success"`
}

// FanOutResult summarizes the gathered outputs across all scattered items.
type FanOutResult struct {
	TotalItems int                `json:"total_items"`
	Successful int                `json:"successful"`
	Failed     int                `json:"failed"`
	Items      []FanOutItemResult `json:"items"`
}

// ItemProcessor is the callback function invoked for each scattered item.
type ItemProcessor func(index int, item interface{}) (json.RawMessage, error)

// DynamicFanOutCoordinator coordinates scatter-gather execution with concurrency bounds.
type DynamicFanOutCoordinator struct {
	cfg FanOutConfig
}

// NewDynamicFanOutCoordinator creates a new scatter-gather coordinator.
func NewDynamicFanOutCoordinator(cfg FanOutConfig) *DynamicFanOutCoordinator {
	if cfg.MaxParallelism <= 0 {
		cfg.MaxParallelism = 1
	}
	if cfg.Policy == "" {
		cfg.Policy = PolicyCollectAll
	}
	return &DynamicFanOutCoordinator{cfg: cfg}
}

// ScatterGather executes processor over items concurrently up to MaxParallelism.
func (c *DynamicFanOutCoordinator) ScatterGather(items []interface{}, processor ItemProcessor) (*FanOutResult, error) {
	n := len(items)
	result := &FanOutResult{
		TotalItems: n,
		Items:      make([]FanOutItemResult, n),
	}

	if n == 0 {
		return result, nil
	}

	sem := make(chan struct{}, c.cfg.MaxParallelism)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	stopExecution := false

	for i, it := range items {
		mu.Lock()
		if stopExecution {
			mu.Unlock()
			break
		}
		mu.Unlock()

		sem <- struct{}{}
		wg.Add(1)

		go func(idx int, item interface{}) {
			defer func() {
				<-sem
				wg.Done()
			}()

			out, err := processor(idx, item)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				result.Failed++
				result.Items[idx] = FanOutItemResult{
					Index:   idx,
					Item:    item,
					Error:   err.Error(),
					Success: false,
				}

				if c.cfg.Policy == PolicyFailFast {
					if firstErr == nil {
						firstErr = err
					}
					stopExecution = true
				}
			} else {
				result.Successful++
				result.Items[idx] = FanOutItemResult{
					Index:   idx,
					Item:    item,
					Output:  out,
					Success: true,
				}
			}
		}(i, it)
	}

	wg.Wait()

	if firstErr != nil && c.cfg.Policy == PolicyFailFast {
		return result, fmt.Errorf("fan-out execution aborted on failure: %w", firstErr)
	}

	if c.cfg.Policy == PolicyCollectAll && result.Failed > 0 {
		return result, fmt.Errorf("fan-out completed with %d/%d item failures", result.Failed, n)
	}

	return result, nil
}
