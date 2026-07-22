package core

import (
	"context"
	"testing"
)

func TestDeadLetterQueueRouter(t *testing.T) {
	router := NewDeadLetterQueueRouter(3)

	for i := 1; i <= 5; i++ {
		err := router.RouteToDLQ(context.Background(), DeadLetterStepRecord{
			WorkflowID:    "wf-fail",
			StepID:        "step-fatal",
			TotalRetries:  5,
			TerminalError: "panic: null pointer dereference",
		})
		if err != nil {
			t.Fatalf("unexpected route error: %v", err)
		}
	}

	if router.Count() != 3 {
		t.Errorf("expected 3 items retained after ring truncation, got %d", router.Count())
	}

	recs := router.GetRecords()
	if len(recs) != 3 {
		t.Errorf("expected 3 records, got %d", len(recs))
	}
}
