package postgres

import (
	"strings"
	"testing"
)

func TestVacuumAdvisor_XIDWraparoundCritical(t *testing.T) {
	advisor := NewVacuumAdvisor(DefaultVacuumAdvisorConfig())

	stats := TableBloatStats{
		TableName:          "events",
		LiveTuples:         1000000,
		DeadTuples:         1000,
		XIDAge:             1800000000,
		AutovacuumXIDLimit: 2000000000, // 90% reached
	}

	rec := advisor.EvaluateTable(stats)
	if rec.Urgency != "CRITICAL" {
		t.Errorf("expected CRITICAL urgency, got %s", rec.Urgency)
	}
	if rec.Action != ActionVacuumFreeze {
		t.Errorf("expected VACUUM FREEZE, got %s", rec.Action)
	}
	if !strings.Contains(rec.SuggestedSQL, "VACUUM (FREEZE, VERBOSE)") {
		t.Errorf("unexpected SQL: %s", rec.SuggestedSQL)
	}
}

func TestVacuumAdvisor_HighBloat(t *testing.T) {
	advisor := NewVacuumAdvisor(DefaultVacuumAdvisorConfig())

	stats := TableBloatStats{
		TableName:  "temp_runs",
		LiveTuples: 20000,
		DeadTuples: 30000, // 60% dead
	}

	rec := advisor.EvaluateTable(stats)
	if rec.Action != ActionVacuumFull {
		t.Errorf("expected ActionVacuumFull, got %s", rec.Action)
	}
	if rec.Urgency != "HIGH" {
		t.Errorf("expected HIGH, got %s", rec.Urgency)
	}
}

func TestVacuumAdvisor_Healthy(t *testing.T) {
	advisor := NewVacuumAdvisor(DefaultVacuumAdvisorConfig())

	stats := TableBloatStats{
		TableName:  "orders",
		LiveTuples: 100000,
		DeadTuples: 2000, // 2% dead
	}

	rec := advisor.EvaluateTable(stats)
	if rec.Action != ActionNone {
		t.Errorf("expected ActionNone for healthy table, got %s", rec.Action)
	}
}
