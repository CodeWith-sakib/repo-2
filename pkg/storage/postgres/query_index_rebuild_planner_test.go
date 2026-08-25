package postgres

import (
	"testing"
)

func TestIndexRebuildPlanner(t *testing.T) {
	planner := NewIndexRebuildPlanner()

	t1 := planner.ScheduleReindex("kf_workflows", "idx_wf_created", 50.0)
	if t1.Priority != 3 {
		t.Errorf("expected priority 3 for 50MB bloat, got %d", t1.Priority)
	}

	t2 := planner.ScheduleReindex("kf_step_runs", "idx_steps_status", 1500.0)
	if t2.Priority != 1 {
		t.Errorf("expected priority 1 for 1500MB bloat, got %d", t2.Priority)
	}

	tasks := planner.PendingTasks()
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	planner.Clear()
	if len(planner.PendingTasks()) != 0 {
		t.Error("expected empty pending tasks after clear")
	}
}
