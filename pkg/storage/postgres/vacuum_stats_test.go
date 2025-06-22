package postgres

import "testing"

func TestVacuumTelemetry(t *testing.T) {
	vt := NewVacuumTelemetry()
	vt.Record(50)
	vt.Record(25)
	if vt.TuplesPruned != 75 {
		t.Errorf("expected 75, got %d", vt.TuplesPruned)
	}
}
