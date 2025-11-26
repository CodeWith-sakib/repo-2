package postgres

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// PaginationDirection defines cursor seek traversal direction.
type PaginationDirection string

const (
	SeekForward  PaginationDirection = "FORWARD"
	SeekBackward PaginationDirection = "BACKWARD"
)

// KeysetCursor stores encoded opaque pagination seek points.
type KeysetCursor struct {
	SortValue interface{}         `json:"sort_val"`
	ID        string              `json:"id"`
	Direction PaginationDirection `json:"dir"`
}

// EncodeCursor serializes a cursor into URL-safe base64.
func EncodeCursor(sortVal interface{}, id string, dir PaginationDirection) (string, error) {
	c := KeysetCursor{
		SortValue: sortVal,
		ID:        id,
		Direction: dir,
	}
	bytes, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// DecodeCursor decodes a URL-safe base64 string into a KeysetCursor.
func DecodeCursor(encoded string) (*KeysetCursor, error) {
	bytes, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor base64: %w", err)
	}
	var c KeysetCursor
	if err := json.Unmarshal(bytes, &c); err != nil {
		return nil, fmt.Errorf("invalid cursor json: %w", err)
	}
	return &c, nil
}

// KeysetPaginationBuilder builds seek-based SQL pagination queries.
type KeysetPaginationBuilder struct {
	table      string
	sortColumn string
	idColumn   string
	descending bool
	pageSize   int
}

// NewKeysetPaginationBuilder creates a keyset pagination builder.
func NewKeysetPaginationBuilder(table, sortCol, idCol string, descending bool, pageSize int) (*KeysetPaginationBuilder, error) {
	if table == "" || sortCol == "" || idCol == "" {
		return nil, fmt.Errorf("table, sortCol, and idCol cannot be empty")
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return &KeysetPaginationBuilder{
		table:      table,
		sortColumn: sortCol,
		idColumn:   idCol,
		descending: descending,
		pageSize:   pageSize,
	}, nil
}

// BuildQuery constructs the keyset query with parameter placeholders and arguments.
func (b *KeysetPaginationBuilder) BuildQuery(cursor *KeysetCursor, baseWhere string, startArgIdx int) (string, []interface{}) {
	var sb strings.Builder
	sb.WriteString("SELECT * FROM ")
	sb.WriteString(b.table)

	var conditions []string
	if baseWhere != "" {
		conditions = append(conditions, baseWhere)
	}

	var args []interface{}
	idx := startArgIdx

	if cursor != nil {
		op := ">"
		if b.descending {
			op = "<"
		}
		if cursor.Direction == SeekBackward {
			// Invert operator for backward traversal
			if op == "<" {
				op = ">"
			} else {
				op = "<"
			}
		}

		// Composite row comparison: (sort_col, id) < ($1, $2)
		clause := fmt.Sprintf("(%s, %s) %s ($%d, $%d)", b.sortColumn, b.idColumn, op, idx, idx+1)
		conditions = append(conditions, clause)

		args = append(args, cursor.SortValue, cursor.ID)
		idx += 2
	}

	if len(conditions) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(conditions, " AND "))
	}

	// Order by clause
	dirStr := "ASC"
	if b.descending {
		dirStr = "DESC"
	}
	sb.WriteString(fmt.Sprintf(" ORDER BY %s %s, %s %s", b.sortColumn, dirStr, b.idColumn, dirStr))

	// Fetch limit + 1 to detect if next page exists
	sb.WriteString(fmt.Sprintf(" LIMIT %d", b.pageSize+1))

	return sb.String(), args
}

// FormatTimestampForCursor converts time.Time to RFC3339 for cursor serialization.
func FormatTimestampForCursor(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
