package postgres

import (
	"context"
	"fmt"
	"sync"
)

// SavepointState describes the lifecycle of a savepoint within a DB transaction.
type SavepointState string

const (
	SavepointActive     SavepointState = "active"
	SavepointReleased   SavepointState = "released"
	SavepointRolledBack SavepointState = "rolled_back"
)

// Savepoint represents a named nested transaction checkpoint.
type Savepoint struct {
	Name  string
	State SavepointState
}

// SavepointMachine manages an ordered stack of savepoints within a database transaction,
// supporting nested rollback and release semantics.
type SavepointMachine struct {
	mu     sync.Mutex
	stack  []*Savepoint
	txID   string
	execFn func(ctx context.Context, sql string) error
}

// NewSavepointMachine creates a machine tied to a logical transaction ID with a query executor.
func NewSavepointMachine(txID string, execFn func(ctx context.Context, sql string) error) *SavepointMachine {
	return &SavepointMachine{
		txID:   txID,
		execFn: execFn,
	}
}

// Create establishes a new named savepoint and pushes it on the stack.
func (m *SavepointMachine) Create(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, sp := range m.stack {
		if sp.Name == name && sp.State == SavepointActive {
			return fmt.Errorf("savepoint %q already active in tx %s", name, m.txID)
		}
	}

	if err := m.execFn(ctx, fmt.Sprintf("SAVEPOINT %s", name)); err != nil {
		return fmt.Errorf("failed to create savepoint %q: %w", name, err)
	}

	m.stack = append(m.stack, &Savepoint{Name: name, State: SavepointActive})
	return nil
}

// Release releases the named savepoint, making its changes permanent within the outer transaction.
func (m *SavepointMachine) Release(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sp, err := m.findActive(name)
	if err != nil {
		return err
	}

	if err := m.execFn(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", name)); err != nil {
		return fmt.Errorf("failed to release savepoint %q: %w", name, err)
	}

	sp.State = SavepointReleased

	// Also mark any nested savepoints created after this one as released
	m.cascadeRelease(name)
	return nil
}

// RollbackTo rolls back the transaction to the named savepoint without aborting the outer tx.
func (m *SavepointMachine) RollbackTo(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sp, err := m.findActive(name)
	if err != nil {
		return err
	}

	if err := m.execFn(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", name)); err != nil {
		return fmt.Errorf("failed to rollback to savepoint %q: %w", name, err)
	}

	sp.State = SavepointRolledBack

	// Invalidate all savepoints after this one in stack order
	m.invalidateAfter(name)
	return nil
}

// Depth returns the number of currently active savepoints on the stack.
func (m *SavepointMachine) Depth() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, sp := range m.stack {
		if sp.State == SavepointActive {
			count++
		}
	}
	return count
}

func (m *SavepointMachine) findActive(name string) (*Savepoint, error) {
	for i := len(m.stack) - 1; i >= 0; i-- {
		if m.stack[i].Name == name && m.stack[i].State == SavepointActive {
			return m.stack[i], nil
		}
	}
	return nil, fmt.Errorf("no active savepoint named %q in tx %s", name, m.txID)
}

func (m *SavepointMachine) cascadeRelease(name string) {
	found := false
	for _, sp := range m.stack {
		if sp.Name == name {
			found = true
			continue
		}
		if found && sp.State == SavepointActive {
			sp.State = SavepointReleased
		}
	}
}

func (m *SavepointMachine) invalidateAfter(name string) {
	found := false
	for _, sp := range m.stack {
		if sp.Name == name {
			found = true
			continue
		}
		if found && sp.State == SavepointActive {
			sp.State = SavepointRolledBack
		}
	}
}
