package postgres

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"
)

// MigrationDirection represents migration direction.
type MigrationDirection string

const (
	DirectionUp   MigrationDirection = "UP"
	DirectionDown MigrationDirection = "DOWN"
)

// MigrationScript contains raw SQL scripts for schema alterations.
type MigrationScript struct {
	Version int64
	Name    string
	UpSQL   string
	DownSQL string
}

// Checksum returns the SHA-256 hash of the UpSQL script.
func (s *MigrationScript) Checksum() string {
	h := sha256.Sum256([]byte(s.UpSQL))
	return hex.EncodeToString(h[:])
}

// AppliedMigration records a migration applied to the database.
type AppliedMigration struct {
	Version   int64
	Name      string
	Checksum  string
	AppliedAt time.Time
}

// MigrationPlan represents the ordered steps to apply.
type MigrationPlan struct {
	Direction MigrationDirection
	Scripts   []*MigrationScript
}

// SchemaMigrationCoordinator orchestrates migration version planning and integrity checks.
type SchemaMigrationCoordinator struct {
	registry map[int64]*MigrationScript
}

// NewSchemaMigrationCoordinator creates a new migration coordinator.
func NewSchemaMigrationCoordinator() *SchemaMigrationCoordinator {
	return &SchemaMigrationCoordinator{
		registry: make(map[int64]*MigrationScript),
	}
}

// Register adds a migration script to the coordinator.
func (c *SchemaMigrationCoordinator) Register(script *MigrationScript) error {
	if script == nil || script.Version <= 0 {
		return fmt.Errorf("invalid migration version: %v", script)
	}
	if _, exists := c.registry[script.Version]; exists {
		return fmt.Errorf("duplicate migration version: %d", script.Version)
	}
	c.registry[script.Version] = script
	return nil
}

// PlanUp computes pending unapplied migrations in ascending version order.
func (c *SchemaMigrationCoordinator) PlanUp(applied []AppliedMigration) (*MigrationPlan, error) {
	appliedMap := make(map[int64]AppliedMigration, len(applied))
	for _, a := range applied {
		// Integrity check: verify checksum matches registered version
		script, exists := c.registry[a.Version]
		if exists && script.Checksum() != a.Checksum {
			return nil, fmt.Errorf("checksum mismatch for migration %d (%s): registered=%s, applied=%s",
				a.Version, a.Name, script.Checksum(), a.Checksum)
		}
		appliedMap[a.Version] = a
	}

	var pending []*MigrationScript
	for v, s := range c.registry {
		if _, ok := appliedMap[v]; !ok {
			pending = append(pending, s)
		}
	}

	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Version < pending[j].Version
	})

	return &MigrationPlan{
		Direction: DirectionUp,
		Scripts:   pending,
	}, nil
}

// PlanDown computes rollback scripts down to targetVersion in descending order.
func (c *SchemaMigrationCoordinator) PlanDown(applied []AppliedMigration, targetVersion int64) (*MigrationPlan, error) {
	var toRollback []*MigrationScript
	for _, a := range applied {
		if a.Version > targetVersion {
			script, exists := c.registry[a.Version]
			if !exists {
				return nil, fmt.Errorf("cannot rollback unknown migration version %d", a.Version)
			}
			toRollback = append(toRollback, script)
		}
	}

	// Sort descending for rollback
	sort.Slice(toRollback, func(i, j int) bool {
		return toRollback[i].Version > toRollback[j].Version
	})

	return &MigrationPlan{
		Direction: DirectionDown,
		Scripts:   toRollback,
	}, nil
}

// SchemaMigrationsTableDDL returns the standard SQL to create the migration metadata tracking table.
func SchemaMigrationsTableDDL() string {
	return `CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    checksum VARCHAR(64) NOT NULL,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);`
}
