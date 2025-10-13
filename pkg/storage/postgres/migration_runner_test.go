package postgres

import (
	"testing"
	"time"
)

func TestSchemaMigrationCoordinator_PlanUpAndDown(t *testing.T) {
	coord := NewSchemaMigrationCoordinator()

	s1 := &MigrationScript{
		Version: 1,
		Name:    "create_users",
		UpSQL:   "CREATE TABLE users (id SERIAL PRIMARY KEY);",
		DownSQL: "DROP TABLE users;",
	}
	s2 := &MigrationScript{
		Version: 2,
		Name:    "add_email",
		UpSQL:   "ALTER TABLE users ADD COLUMN email VARCHAR(255);",
		DownSQL: "ALTER TABLE users DROP COLUMN email;",
	}
	s3 := &MigrationScript{
		Version: 3,
		Name:    "create_orders",
		UpSQL:   "CREATE TABLE orders (id SERIAL PRIMARY KEY);",
		DownSQL: "DROP TABLE orders;",
	}

	_ = coord.Register(s1)
	_ = coord.Register(s2)
	_ = coord.Register(s3)

	// Currently applied: version 1
	applied := []AppliedMigration{
		{Version: 1, Name: "create_users", Checksum: s1.Checksum(), AppliedAt: time.Now()},
	}

	planUp, err := coord.PlanUp(applied)
	if err != nil {
		t.Fatalf("plan up failed: %v", err)
	}

	if len(planUp.Scripts) != 2 {
		t.Fatalf("expected 2 pending migrations, got %d", len(planUp.Scripts))
	}
	if planUp.Scripts[0].Version != 2 || planUp.Scripts[1].Version != 3 {
		t.Errorf("unexpected script order: %v, %v", planUp.Scripts[0].Version, planUp.Scripts[1].Version)
	}

	// Now test rollback down to version 1 when all 3 are applied
	allApplied := []AppliedMigration{
		{Version: 1, Name: "create_users", Checksum: s1.Checksum(), AppliedAt: time.Now()},
		{Version: 2, Name: "add_email", Checksum: s2.Checksum(), AppliedAt: time.Now()},
		{Version: 3, Name: "create_orders", Checksum: s3.Checksum(), AppliedAt: time.Now()},
	}

	planDown, err := coord.PlanDown(allApplied, 1)
	if err != nil {
		t.Fatalf("plan down failed: %v", err)
	}

	if len(planDown.Scripts) != 2 {
		t.Fatalf("expected 2 rollback migrations, got %d", len(planDown.Scripts))
	}
	// Descending order for rollback: 3 then 2
	if planDown.Scripts[0].Version != 3 || planDown.Scripts[1].Version != 2 {
		t.Errorf("expected rollback order [3, 2], got [%d, %d]", planDown.Scripts[0].Version, planDown.Scripts[1].Version)
	}
}

func TestSchemaMigrationCoordinator_ChecksumMismatch(t *testing.T) {
	coord := NewSchemaMigrationCoordinator()
	s1 := &MigrationScript{
		Version: 1,
		Name:    "init",
		UpSQL:   "CREATE TABLE test (id INT);",
	}
	_ = coord.Register(s1)

	// Tampered checksum in applied table
	applied := []AppliedMigration{
		{Version: 1, Name: "init", Checksum: "tampered_checksum", AppliedAt: time.Now()},
	}

	_, err := coord.PlanUp(applied)
	if err == nil {
		t.Error("expected checksum mismatch error")
	}
}
