package core

import (
	"context"
	"testing"
	"time"
)

func TestForkJoinCoordinator_SuccessAndReduction(t *testing.T) {
	executor := func(ctx context.Context, branch ForkBranchDefinition) (interface{}, error) {
		val := branch.Payload.(int)
		return val * 10, nil
	}

	reducer := func(results []ForkBranchResult) (interface{}, error) {
		sum := 0
		for _, r := range results {
			if r.Error != nil {
				return nil, r.Error
			}
			sum += r.Output.(int)
		}
		return sum, nil
	}

	coord, err := NewForkJoinCoordinator(executor, reducer)
	if err != nil {
		t.Fatalf("create coordinator failed: %v", err)
	}

	branches := []ForkBranchDefinition{
		{BranchID: "b1", Payload: 1},
		{BranchID: "b2", Payload: 2},
		{BranchID: "b3", Payload: 3},
	}

	reduced, results, err := coord.ExecuteForks(context.Background(), branches)
	if err != nil {
		t.Fatalf("execute forks failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Sum should be (1*10) + (2*10) + (3*10) = 60
	if reduced.(int) != 60 {
		t.Errorf("expected reduced sum 60, got %v", reduced)
	}
}

func TestForkJoinCoordinator_BranchTimeout(t *testing.T) {
	executor := func(ctx context.Context, branch ForkBranchDefinition) (interface{}, error) {
		select {
		case <-time.After(100 * time.Millisecond):
			return "done", nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	coord, _ := NewForkJoinCoordinator(executor, nil)

	branches := []ForkBranchDefinition{
		{BranchID: "slow-branch", Timeout: 10 * time.Millisecond},
	}

	_, results, _ := coord.ExecuteForks(context.Background(), branches)
	if len(results) != 1 {
		t.Fatalf("expected 1 result")
	}

	if results[0].Error != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded error, got %v", results[0].Error)
	}
}
