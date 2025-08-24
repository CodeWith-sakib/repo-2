package postgres

import (
	"context"
	"testing"
)

func TestSavepointMachine_CreateReleaseRollback(t *testing.T) {
	var executed []string
	execFn := func(ctx context.Context, sql string) error {
		executed = append(executed, sql)
		return nil
	}

	m := NewSavepointMachine("tx-001", execFn)

	ctx := context.Background()

	if err := m.Create(ctx, "sp1"); err != nil {
		t.Fatalf("create sp1: %v", err)
	}
	if err := m.Create(ctx, "sp2"); err != nil {
		t.Fatalf("create sp2: %v", err)
	}

	if m.Depth() != 2 {
		t.Errorf("expected depth 2, got %d", m.Depth())
	}

	// Rollback to sp1 should invalidate sp2
	if err := m.RollbackTo(ctx, "sp1"); err != nil {
		t.Fatalf("rollback to sp1: %v", err)
	}

	if m.Depth() != 0 {
		t.Errorf("expected depth 0 after rollback, got %d", m.Depth())
	}

	// Create new sp after rollback
	if err := m.Create(ctx, "sp3"); err != nil {
		t.Fatalf("create sp3 after rollback: %v", err)
	}
	if err := m.Release(ctx, "sp3"); err != nil {
		t.Fatalf("release sp3: %v", err)
	}

	if m.Depth() != 0 {
		t.Errorf("expected depth 0 after release, got %d", m.Depth())
	}

	expectedSQL := []string{
		"SAVEPOINT sp1",
		"SAVEPOINT sp2",
		"ROLLBACK TO SAVEPOINT sp1",
		"SAVEPOINT sp3",
		"RELEASE SAVEPOINT sp3",
	}
	for i, want := range expectedSQL {
		if i >= len(executed) || executed[i] != want {
			t.Errorf("step %d: want %q, got %q", i, want, executed[i])
		}
	}
}
