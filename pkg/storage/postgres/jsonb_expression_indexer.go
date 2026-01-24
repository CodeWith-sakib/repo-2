package postgres

import (
	"fmt"
	"strings"
)

// JSONBIndexType specifies index operator class.
type JSONBIndexType string

const (
	JSONBIndexGIN  JSONBIndexType = "GIN"
	JSONBIndexBTREE JSONBIndexType = "BTREE"
)

// JSONBIndexDef describes an extracted expression index on a JSONB column.
type JSONBIndexDef struct {
	TableName string
	ColumnName string
	JSONPath   string // e.g. payload->'user'->>'id'
	DataType   string // e.g. text, int, timestamp
	IndexType  JSONBIndexType
}

// JSONBExpressionIndexer synthesizes optimal DDL for indexing deep JSONB attributes.
type JSONBExpressionIndexer struct{}

// NewJSONBExpressionIndexer creates an indexer helper.
func NewJSONBExpressionIndexer() *JSONBExpressionIndexer {
	return &JSONBExpressionIndexer{}
}

// BuildIndexDDL constructs concurrent PostgreSQL CREATE INDEX DDL statement.
func (idx *JSONBExpressionIndexer) BuildIndexDDL(def JSONBIndexDef) (string, string, error) {
	if def.TableName == "" || def.ColumnName == "" || def.JSONPath == "" {
		return "", "", fmt.Errorf("table, column, and JSONPath are required")
	}

	cleanPath := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(def.JSONPath, "'", ""), "->", "_"), ">", "")
	idxName := fmt.Sprintf("idx_%s_%s_%s", def.TableName, def.ColumnName, cleanPath)

	var ddl string
	if def.IndexType == JSONBIndexGIN {
		ddl = fmt.Sprintf("CREATE INDEX CONCURRENTLY IF NOT EXISTS %s ON %s USING GIN ((%s%s));",
			idxName, def.TableName, def.ColumnName, def.JSONPath)
	} else {
		// BTREE expression index with typecast
		cast := ""
		if def.DataType != "" {
			cast = fmt.Sprintf("::%s", def.DataType)
		}
		ddl = fmt.Sprintf("CREATE INDEX CONCURRENTLY IF NOT EXISTS %s ON %s (((%s%s)%s));",
			idxName, def.TableName, def.ColumnName, def.JSONPath, cast)
	}

	return idxName, ddl, nil
}
