package postgres

import (
	"fmt"
	"strings"
)

// ConflictAction defines the behavior when unique constraints conflict.
type ConflictAction string

const (
	ConflictDoNothing ConflictAction = "DO NOTHING"
	ConflictDoUpdate  ConflictAction = "DO UPDATE"
)

// BatchUpsertConfig specifies the configuration for generating bulk upsert statements.
type BatchUpsertConfig struct {
	Table              string
	Columns            []string
	ConflictColumns    []string
	ConflictConstraint string
	Action             ConflictAction
	UpdateColumns      []string
	WhereClause        string
	ReturningColumns   []string
}

// UpsertBatch generates chunked parameterized queries for bulk insertions with conflict resolution.
type UpsertBatch struct {
	cfg BatchUpsertConfig
}

// NewUpsertBatch initializes a new batch upsert helper.
func NewUpsertBatch(cfg BatchUpsertConfig) (*UpsertBatch, error) {
	if cfg.Table == "" {
		return nil, fmt.Errorf("table name is required")
	}
	if len(cfg.Columns) == 0 {
		return nil, fmt.Errorf("at least one column is required")
	}
	if len(cfg.ConflictColumns) == 0 && cfg.ConflictConstraint == "" {
		return nil, fmt.Errorf("conflict columns or conflict constraint must be specified")
	}
	if cfg.Action == "" {
		cfg.Action = ConflictDoNothing
	}
	if cfg.Action == ConflictDoUpdate && len(cfg.UpdateColumns) == 0 {
		return nil, fmt.Errorf("update columns required when action is DO UPDATE")
	}
	return &UpsertBatch{cfg: cfg}, nil
}

// BuildQuery generates a single SQL query and flattened argument slice for the given rows.
func (b *UpsertBatch) BuildQuery(rows [][]interface{}) (string, []interface{}, error) {
	if len(rows) == 0 {
		return "", nil, fmt.Errorf("cannot build query for empty rows slice")
	}

	colCount := len(b.cfg.Columns)
	for i, row := range rows {
		if len(row) != colCount {
			return "", nil, fmt.Errorf("row %d has %d values, expected %d", i, len(row), colCount)
		}
	}

	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	sb.WriteString(b.cfg.Table)
	sb.WriteString(" (")
	sb.WriteString(strings.Join(b.cfg.Columns, ", "))
	sb.WriteString(") VALUES ")

	var args []interface{}
	argIdx := 1

	for i, row := range rows {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("(")
		for j, val := range row {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("$%d", argIdx))
			argIdx++
			args = append(args, val)
		}
		sb.WriteString(")")
	}

	// Conflict target
	sb.WriteString(" ON CONFLICT ")
	if b.cfg.ConflictConstraint != "" {
		sb.WriteString(fmt.Sprintf("ON CONSTRAINT %s ", b.cfg.ConflictConstraint))
	} else if len(b.cfg.ConflictColumns) > 0 {
		sb.WriteString("(")
		sb.WriteString(strings.Join(b.cfg.ConflictColumns, ", "))
		sb.WriteString(") ")
	}

	// Action
	sb.WriteString(string(b.cfg.Action))

	if b.cfg.Action == ConflictDoUpdate {
		sb.WriteString(" SET ")
		updateSets := make([]string, len(b.cfg.UpdateColumns))
		for i, col := range b.cfg.UpdateColumns {
			updateSets[i] = fmt.Sprintf("%s = EXCLUDED.%s", col, col)
		}
		sb.WriteString(strings.Join(updateSets, ", "))

		if b.cfg.WhereClause != "" {
			sb.WriteString(" WHERE ")
			sb.WriteString(b.cfg.WhereClause)
		}
	}

	// Returning
	if len(b.cfg.ReturningColumns) > 0 {
		sb.WriteString(" RETURNING ")
		sb.WriteString(strings.Join(b.cfg.ReturningColumns, ", "))
	}

	return sb.String(), args, nil
}

// ChunkRows splits a large slice of rows into manageable chunks.
func ChunkRows(rows [][]interface{}, chunkSize int) [][][]interface{} {
	if chunkSize <= 0 || len(rows) == 0 {
		return [][][]interface{}{rows}
	}

	var chunks [][][]interface{}
	for i := 0; i < len(rows); i += chunkSize {
		end := i + chunkSize
		if end > len(rows) {
			end = len(rows)
		}
		chunks = append(chunks, rows[i:end])
	}
	return chunks
}
