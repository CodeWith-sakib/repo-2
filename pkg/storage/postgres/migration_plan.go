package postgres

import (
	"fmt"
	"strings"
)

type ColumnSpec struct {
	Name     string
	Type     string
	Nullable bool
	Default  string
}

type TableSpec struct {
	Name        string
	Columns     map[string]ColumnSpec
	PrimaryKeys []string
}

type SchemaDiff struct {
	AddedTables   []TableSpec
	DroppedTables []string
	AddedColumns  map[string][]ColumnSpec
}

func ComputeSchemaDiff(current, desired map[string]TableSpec) *SchemaDiff {
	diff := &SchemaDiff{
		AddedTables:  make([]TableSpec, 0),
		AddedColumns: make(map[string][]ColumnSpec),
	}

	for name, desTable := range desired {
		curTable, exists := current[name]
		if !exists {
			diff.AddedTables = append(diff.AddedTables, desTable)
			continue
		}

		var newCols []ColumnSpec
		for colName, colSpec := range desTable.Columns {
			if _, colExists := curTable.Columns[colName]; !colExists {
				newCols = append(newCols, colSpec)
			}
		}
		if len(newCols) > 0 {
			diff.AddedColumns[name] = newCols
		}
	}

	for name := range current {
		if _, exists := desired[name]; !exists {
			diff.DroppedTables = append(diff.DroppedTables, name)
		}
	}

	return diff
}

func (diff *SchemaDiff) GenerateDDL() []string {
	var stmts []string
	for _, t := range diff.AddedTables {
		var colDefs []string
		for _, col := range t.Columns {
			nullStr := "NOT NULL"
			if col.Nullable {
				nullStr = "NULL"
			}
			colDefs = append(colDefs, fmt.Sprintf("%s %s %s", col.Name, col.Type, nullStr))
		}
		stmts = append(stmts, fmt.Sprintf("CREATE TABLE %s (%s);", t.Name, strings.Join(colDefs, ", ")))
	}

	for table, cols := range diff.AddedColumns {
		for _, col := range cols {
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", table, col.Name, col.Type))
		}
	}

	for _, t := range diff.DroppedTables {
		stmts = append(stmts, fmt.Sprintf("DROP TABLE %s CASCADE;", t))
	}
	return stmts
}
