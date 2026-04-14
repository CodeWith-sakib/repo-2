package core

import (
	"context"
	"fmt"
	"sync"
)

// ScatterItem represents an individual partitioned element to process in parallel.
type ScatterItem struct {
	Index int         `json:"index"`
	Item  interface{} `json:"item"`
}

// GatherResult holds aggregated outcomes of all scattered partition executions.
type GatherResult struct {
	TotalProcessed int           `json:"total_processed"`
	SuccessCount   int           `json:"success_count"`
	FailureCount   int           `json:"failure_count"`
	Results        []interface{} `json:"results"`
	Errors         []string      `json:"errors"`
}

// ScatterGatherWorker defines worker task processor for a single item.
type ScatterGatherWorker func(ctx context.Context, item ScatterItem) (interface{}, error)

// WorkflowScatterGatherExecutor coordinates map-reduce style scatter/gather operations across worker pools.
type WorkflowScatterGatherExecutor struct {
	maxConcurrency int
}

// NewWorkflowScatterGatherExecutor creates a scatter/gather coordinator.
func NewWorkflowScatterGatherExecutor(concurrency int) *WorkflowScatterGatherExecutor {
	if concurrency <= 0 {
		concurrency = 10
	}
	return &WorkflowScatterGatherExecutor{
		maxConcurrency: concurrency,
	}
}

// Execute fans out processing across items and gathers combined results.
func (e *WorkflowScatterGatherExecutor) Execute(ctx context.Context, items []interface{}, worker ScatterGatherWorker) (*GatherResult, error) {
	if worker == nil {
		return nil, fmt.Errorf("worker task cannot be nil")
	}

	result := &GatherResult{
		TotalProcessed: len(items),
		Results:        make([]interface{}, len(items)),
		Errors:         make([]string, 0),
	}

	if len(items) == 0 {
		return result, nil
	}

	sem := make(chan struct{}, e.maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, it := range items {
		wg.Add(1)
		go func(idx int, val interface{}) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			res, err := worker(ctx, ScatterItem{Index: idx, Item: val})
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				result.FailureCount++
				result.Errors = append(result.Errors, fmt.Sprintf("item %d: %v", idx, err))
			} else {
				result.SuccessCount++
				result.Results[idx] = res
			}
		}(i, it)
	}

	wg.Wait()
	return result, nil
}
