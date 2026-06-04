package postgres

import (
	"context"
	"testing"
)

func TestVacuumFreezeMonitor(t *testing.T) {
	monitor := NewVacuumFreezeMonitor(100_000, 0.70)

	s1 := monitor.RecordTableAge(context.Background(), "kf_workflows", 50_000)
	if s1.RequiresUrgentFreeze {
		t.Errorf("50,000 age should not require urgent freeze")
	}

	s2 := monitor.RecordTableAge(context.Background(), "kf_step_runs", 75_000)
	if !s2.RequiresUrgentFreeze {
		t.Errorf("75,000 age should require urgent freeze")
	}

	urgent := monitor.GetUrgentTables()
	if len(urgent) != 1 || urgent[0].TableName != "kf_step_runs" {
		t.Errorf("expected kf_step_runs to be urgent, got %v", urgent)
	}
}
