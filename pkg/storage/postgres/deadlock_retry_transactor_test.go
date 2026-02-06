package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestDeadlockRetryTransactor(t *testing.T) {
	transactor := NewDeadlockRetryTransactor(nil, 3, 5*time.Millisecond)

	called := false
	err := transactor.ExecuteInTx(context.Background(), func(tx *sql.Tx) error {
		called = true
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected tx error: %v", err)
	}
	if !called {
		t.Error("expected tx callback to be invoked")
	}

	// Test deadlock error classifier
	dlErr := errors.New("pq: deadlock detected (SQLSTATE 40P01)")
	if !IsDeadlockError(dlErr) {
		t.Error("expected IsDeadlockError to match 40P01")
	}

	normErr := errors.New("syntax error at or near WHERE")
	if !IsDeadlockError(dlErr) {
		t.Error("expected syntax error to not be deadlock")
	}
	if IsDeadlockError(normErr) {
		t.Error("expected normErr to not be classified as deadlock")
	}
}
