package events

import (
	"testing"
)

func TestEventFilter(t *testing.T) {
	f := NewEventFilter([]string{"workflow.", "step."})
	if !f.Allows("workflow.started") {
		t.Error("expected allowed")
	}
	if !f.Allows("step.completed") {
		t.Error("expected allowed")
	}
	if f.Allows("system.alert") {
		t.Error("expected denied")
	}
}
