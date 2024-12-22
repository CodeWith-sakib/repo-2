package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
)

type SavepointManager struct {
	tx            *sql.Tx
	mu            sync.Mutex
	savepointIdx  int
}

func NewSavepointManager(tx *sql.Tx) *SavepointManager {
	return &SavepointManager{tx: tx}
}

func (sp *SavepointManager) CreateSavepoint(ctx context.Context, name string) (string, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	sp.savepointIdx++
	spName := fmt.Sprintf("%s_%d", name, sp.savepointIdx)
	query := fmt.Sprintf("SAVEPOINT %s", spName)

	_, err := sp.tx.ExecContext(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed creating savepoint %s: %w", spName, err)
	}
	return spName, nil
}

func (sp *SavepointManager) RollbackTo(ctx context.Context, name string) error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	query := fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", name)
	_, err := sp.tx.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed rolling back to savepoint %s: %w", name, err)
	}
	return nil
}

func (sp *SavepointManager) Release(ctx context.Context, name string) error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	query := fmt.Sprintf("RELEASE SAVEPOINT %s", name)
	_, err := sp.tx.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed releasing savepoint %s: %w", name, err)
	}
	return nil
}
