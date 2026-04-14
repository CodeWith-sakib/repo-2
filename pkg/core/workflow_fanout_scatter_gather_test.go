package core

import (
	"context"
	"fmt"
	"testing"
)

func TestWorkflowScatterGatherExecutor(t *testing.T) {
	executor := NewWorkflowScatterGatherExecutor(4)

	items := []interface{}{10, 20, 30, 40, 50}
	worker := func(ctx context.Context, item ScatterItem) (interface{}, error) {
		val := item.Item.(int)
		if val == 30 {
			return nil, fmt.Errorf("simulated error on 30")
		}
		return val * 2, nil
	}

	res, err := executor.Execute(context.Background(), items, worker)
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}

	if res.TotalProcessed != 5 {
		t.Errorf("expected 5 processed, got %d", res.TotalProcessed)
	}
	if res.SuccessCount != 4 {
		t.Errorf("expected 4 successes, got %d", res.SuccessCount)
	}
	if res.FailureCount != 1 {
		t.Errorf("expected 1 failure, got %d", res.FailureCount)
	}
	if res.Results[0].(int) != 20 || res.Results[1].(int) != 40 {
		t.Errorf("unexpected mapped result values: %v", res.Results)
	}
}
