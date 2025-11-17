package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ForkBranchDefinition specifies a sub-branch to be executed concurrently.
type ForkBranchDefinition struct {
	BranchID string
	Payload  interface{}
	Timeout  time.Duration
}

// ForkBranchResult records the execution outcome of an individual parallel branch.
type ForkBranchResult struct {
	BranchID string
	Output   interface{}
	Error    error
	Duration time.Duration
}

// ForkJoinReducer aggregates multiple branch results into a final combined output.
type ForkJoinReducer func(results []ForkBranchResult) (interface{}, error)

// BranchExecutor executes a single fork branch.
type BranchExecutor func(ctx context.Context, branch ForkBranchDefinition) (interface{}, error)

// ForkJoinCoordinator manages structured fork-join parallel branch execution and barrier synchronization.
type ForkJoinCoordinator struct {
	executor BranchExecutor
	reducer  ForkJoinReducer
}

// NewForkJoinCoordinator creates a coordinator with the given executor and reducer.
func NewForkJoinCoordinator(executor BranchExecutor, reducer ForkJoinReducer) (*ForkJoinCoordinator, error) {
	if executor == nil {
		return nil, fmt.Errorf("branch executor cannot be nil")
	}
	return &ForkJoinCoordinator{
		executor: executor,
		reducer:  reducer,
	}, nil
}

// ExecuteForks launches all branches in parallel, waits at the join barrier, and invokes the reducer.
func (c *ForkJoinCoordinator) ExecuteForks(parentCtx context.Context, branches []ForkBranchDefinition) (interface{}, []ForkBranchResult, error) {
	n := len(branches)
	results := make([]ForkBranchResult, n)

	if n == 0 {
		if c.reducer != nil {
			reduced, err := c.reducer(results)
			return reduced, results, err
		}
		return nil, results, nil
	}

	var wg sync.WaitGroup
	wg.Add(n)

	for i, b := range branches {
		go func(idx int, branch ForkBranchDefinition) {
			defer wg.Done()

			start := time.Now()
			branchCtx := parentCtx
			var cancel context.CancelFunc

			if branch.Timeout > 0 {
				branchCtx, cancel = context.WithTimeout(parentCtx, branch.Timeout)
				defer cancel()
			}

			out, err := c.executor(branchCtx, branch)
			results[idx] = ForkBranchResult{
				BranchID: branch.BranchID,
				Output:   out,
				Error:    err,
				Duration: time.Since(start),
			}
		}(i, b)
	}

	wg.Wait()

	if c.reducer != nil {
		reduced, err := c.reducer(results)
		return reduced, results, err
	}

	return nil, results, nil
}
