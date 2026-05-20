package core

import (
	"context"
	"testing"
)

func TestCompensatingActionEngine(t *testing.T) {
	engine := NewCompensatingActionEngine()

	engine.RegisterAction(CompensatingAction{StepID: "reserve_inventory", ActionName: "release_inventory"})
	engine.RegisterAction(CompensatingAction{StepID: "charge_credit_card", ActionName: "refund_payment"})

	if engine.ActionCount() != 2 {
		t.Fatalf("expected 2 actions, got %d", engine.ActionCount())
	}

	var rollbackOrder []string
	handler := func(ctx context.Context, action CompensatingAction) error {
		rollbackOrder = append(rollbackOrder, action.ActionName)
		return nil
	}

	executed, err := engine.Rollback(context.Background(), handler)
	if err != nil {
		t.Fatalf("unexpected rollback error: %v", err)
	}

	if len(executed) != 2 {
		t.Errorf("expected 2 executed rollbacks, got %d", len(executed))
	}

	// Verify LIFO order: refund_payment first, release_inventory second
	if rollbackOrder[0] != "refund_payment" || rollbackOrder[1] != "release_inventory" {
		t.Errorf("expected LIFO rollback order, got %v", rollbackOrder)
	}
}
