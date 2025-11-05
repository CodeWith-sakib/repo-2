package postgres

import (
	"encoding/json"
	"fmt"
	"strings"
)

// GINIndexType specifies the GIN operator class.
type GINIndexType string

const (
	GINDefault GINIndexType = "jsonb_ops"      // default, supports ?, ?|, ?&, @>
	GINPathOps GINIndexType = "jsonb_path_ops" // smaller, faster, supports @>
)

// JSONBIndexBuilder generates DDL statements for PostgreSQL JSONB indexing.
type JSONBIndexBuilder struct {
	table  string
	column string
}

// NewJSONBIndexBuilder initializes an index generator for a given table and JSONB column.
func NewJSONBIndexBuilder(table, column string) (*JSONBIndexBuilder, error) {
	if table == "" || column == "" {
		return nil, fmt.Errorf("table and column cannot be empty")
	}
	return &JSONBIndexBuilder{table: table, column: column}, nil
}

// CreateGINIndex generates DDL for a GIN index on the whole JSONB column.
func (b *JSONBIndexBuilder) CreateGINIndex(indexName string, opClass GINIndexType) string {
	if indexName == "" {
		indexName = fmt.Sprintf("idx_%s_%s_gin", b.table, b.column)
	}
	if opClass == "" {
		opClass = GINDefault
	}
	return fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s USING gin (%s %s);",
		indexName, b.table, b.column, opClass)
}

// CreatePathIndex generates an expression B-tree index on a specific extracted JSON path.
func (b *JSONBIndexBuilder) CreatePathIndex(indexName, path, castType string) string {
	if indexName == "" {
		cleanPath := strings.ReplaceAll(path, ".", "_")
		indexName = fmt.Sprintf("idx_%s_%s", b.table, cleanPath)
	}

	expr := b.buildExtractExpr(path)
	if castType != "" {
		expr = fmt.Sprintf("(%s)::%s", expr, castType)
	}

	return fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s ((%s));",
		indexName, b.table, expr)
}

// BuildContainmentPredicate generates a `@>` JSON containment WHERE condition.
func (b *JSONBIndexBuilder) BuildContainmentPredicate(matchData map[string]interface{}) (string, error) {
	jsonBytes, err := json.Marshal(matchData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal containment data: %w", err)
	}
	// Escape single quotes for SQL literal
	escaped := strings.ReplaceAll(string(jsonBytes), "'", "''")
	return fmt.Sprintf("%s @> '%s'::jsonb", b.column, escaped), nil
}

// BuildKeyExistsPredicate generates a `?` key existence WHERE condition.
func (b *JSONBIndexBuilder) BuildKeyExistsPredicate(key string) string {
	escaped := strings.ReplaceAll(key, "'", "''")
	return fmt.Sprintf("%s ? '%s'", b.column, escaped)
}

func (b *JSONBIndexBuilder) buildExtractExpr(path string) string {
	parts := strings.Split(path, ".")
	if len(parts) == 1 {
		return fmt.Sprintf("%s->>'%s'", b.column, parts[0])
	}

	var sb strings.Builder
	sb.WriteString(b.column)
	for i := 0; i < len(parts)-1; i++ {
		sb.WriteString(fmt.Sprintf("->'%s'", parts[i]))
	}
	sb.WriteString(fmt.Sprintf("->>'%s'", parts[len(parts)-1]))
	return sb.String()
}
