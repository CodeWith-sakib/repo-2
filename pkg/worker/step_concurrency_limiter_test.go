package worker

import (
	"testing"
)

func TestTaskConcurrencyGovernor(t *testing.T) {
	gov := NewTaskConcurrencyGovernor(2)

	if !gov.TryAcquire() {
		t.Fatal("expected 1st acquire to succeed")
	}
	if !gov.TryAcquire() {
		t.Fatal("expected 2nd acquire to succeed")
	}
	// 3rd rejected
	if gov.TryAcquire() {
		t.Error("expected 3rd acquire to fail")
	}

	if gov.CurrentInFlight() != 2 {
		t.Errorf("expected 2 in flight, got %d", gov.CurrentInFlight())
	}

	gov.Release()
	if gov.CurrentInFlight() != 1 {
		t.Errorf("expected 1 in flight after release, got %d", gov.CurrentInFlight())
	}
}
