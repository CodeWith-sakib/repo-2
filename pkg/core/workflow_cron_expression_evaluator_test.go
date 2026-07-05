package core

import (
	"testing"
	"time"
)

func TestCronExpressionEvaluator(t *testing.T) {
	eval := NewCronExpressionEvaluator()

	spec, err := eval.Parse("*/15 2 * * *")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// From 02:00 -> next is 02:15
	ref := time.Date(2026, 6, 5, 2, 0, 0, 0, time.UTC)
	next := eval.NextTrigger(spec, ref)

	expected := time.Date(2026, 6, 5, 2, 15, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, next)
	}

	// Invalid spec
	_, err = eval.Parse("invalid cron line")
	if err == nil {
		t.Error("expected error for invalid spec")
	}
}
