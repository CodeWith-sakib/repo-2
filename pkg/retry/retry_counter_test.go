package retry

import (
	"testing"
)

func TestRetryCounter(t *testing.T) {
	c := NewRetryCounter()
	if c.Inc("task-1") != 1 {
		t.Errorf("expected 1")
	}
	if c.Inc("task-1") != 2 {
		t.Errorf("expected 2")
	}
	if c.Get("task-1") != 2 {
		t.Errorf("expected 2")
	}
	c.Reset("task-1")
	if c.Get("task-1") != 0 {
		t.Errorf("expected 0")
	}
}
