package worker

import (
	"testing"
)

func TestDynamicCapacityManager(t *testing.T) {
	m := NewDynamicCapacityManager(5)
	if m.GetCapacity() != 5 {
		t.Errorf("expected 5, got %d", m.GetCapacity())
	}
	m.SetCapacity(10)
	if m.GetCapacity() != 10 {
		t.Errorf("expected 10, got %d", m.GetCapacity())
	}
}
