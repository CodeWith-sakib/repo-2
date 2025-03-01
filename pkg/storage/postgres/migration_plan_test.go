package postgres

import (
	"testing"
)

func TestSchemaDiffAndDDL(t *testing.T) {
	current := map[string]TableSpec{
		"users": {
			Name: "users",
			Columns: map[string]ColumnSpec{
				"id": {Name: "id", Type: "VARCHAR(64)", Nullable: false},
			},
		},
	}

	desired := map[string]TableSpec{
		"users": {
			Name: "users",
			Columns: map[string]ColumnSpec{
				"id":    {Name: "id", Type: "VARCHAR(64)", Nullable: false},
				"email": {Name: "email", Type: "TEXT", Nullable: true},
			},
		},
		"tasks": {
			Name: "tasks",
			Columns: map[string]ColumnSpec{
				"task_id": {Name: "task_id", Type: "VARCHAR(64)", Nullable: false},
			},
		},
	}

	diff := ComputeSchemaDiff(current, desired)
	if len(diff.AddedTables) != 1 || diff.AddedTables[0].Name != "tasks" {
		t.Fatalf("expected 1 added table 'tasks', got %v", diff.AddedTables)
	}

	if len(diff.AddedColumns["users"]) != 1 {
		t.Fatalf("expected 1 added column for users, got %v", diff.AddedColumns)
	}

	ddl := diff.GenerateDDL()
	if len(ddl) != 2 {
		t.Fatalf("expected 2 DDL statements, got %d: %v", len(ddl), ddl)
	}
}
