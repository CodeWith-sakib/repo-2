package postgres

import (
	"strings"
	"testing"
	"time"
)

func TestDeadlockDetector_CycleAndVictim(t *testing.T) {
	d := NewDeadlockDetector()

	now := time.Now()
	// T1 started earlier, T2 started later
	d.RegisterTx("T1", now.Add(-5*time.Minute))
	d.RegisterTx("T2", now.Add(-1*time.Minute))

	// T1 waits on lock A held by T2
	d.AddWaitEdge("T1", "T2", "lock-A")

	// T2 waits on lock B held by T1 -> circular wait!
	d.AddWaitEdge("T2", "T1", "lock-B")

	cycles := d.DetectCycles()
	if len(cycles) == 0 {
		t.Fatal("expected deadlock cycle to be detected")
	}

	c := cycles[0]
	if len(c.Transactions) != 2 {
		t.Errorf("expected 2 transactions in cycle, got %v", c.Transactions)
	}

	// T2 is younger than T1 -> T2 should be selected as victim
	if c.VictimTx != "T2" {
		t.Errorf("expected T2 as victim, got %s", c.VictimTx)
	}

	if !strings.Contains(c.FormatCycle(), "recommended victim to abort: T2") {
		t.Errorf("unexpected cycle format: %s", c.FormatCycle())
	}
}

func TestDeadlockDetector_NoCycle(t *testing.T) {
	d := NewDeadlockDetector()

	// Linear wait: T1 -> T2 -> T3
	d.AddWaitEdge("T1", "T2", "l1")
	d.AddWaitEdge("T2", "T3", "l2")

	cycles := d.DetectCycles()
	if len(cycles) != 0 {
		t.Errorf("expected 0 cycles for linear wait, got %d", len(cycles))
	}
}
