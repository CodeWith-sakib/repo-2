package postgres

import (
	"fmt"
	"strings"
)

// WindowFrame represents window frame specifications (e.g., ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW).
type WindowFrame struct {
	Mode  string // ROWS or RANGE
	Start string
	End   string
}

// WindowSpec represents an SQL OVER clause.
type WindowSpec struct {
	Name        string
	PartitionBy []string
	OrderBy     []string
	Frame       *WindowFrame
}

// AdvancedSelectBuilder constructs complex SQL queries involving CTEs and Window Functions.
type AdvancedSelectBuilder struct {
	ctes        []string
	table       string
	projections []string
	windows     map[string]WindowSpec
	whereClause []string
	groupBy     []string
	having      []string
	orderBy     []string
	limitVal    int
	offsetVal   int
	args        []interface{}
}

// NewAdvancedSelectBuilder initializes a builder.
func NewAdvancedSelectBuilder() *AdvancedSelectBuilder {
	return &AdvancedSelectBuilder{
		windows: make(map[string]WindowSpec),
	}
}

// WithCTE adds a Common Table Expression to the query.
func (b *AdvancedSelectBuilder) WithCTE(name string, query string) *AdvancedSelectBuilder {
	b.ctes = append(b.ctes, fmt.Sprintf("%s AS (%s)", name, query))
	return b
}

// From sets the base table.
func (b *AdvancedSelectBuilder) From(table string) *AdvancedSelectBuilder {
	b.table = table
	return b
}

// Select adds projection expressions.
func (b *AdvancedSelectBuilder) Select(columns ...string) *AdvancedSelectBuilder {
	b.projections = append(b.projections, columns...)
	return b
}

// Window defines a named window specification.
func (b *AdvancedSelectBuilder) Window(name string, partitionBy []string, orderBy []string) *AdvancedSelectBuilder {
	b.windows[name] = WindowSpec{
		Name:        name,
		PartitionBy: partitionBy,
		OrderBy:     orderBy,
	}
	return b
}

// Where adds filter predicates.
func (b *AdvancedSelectBuilder) Where(condition string, arg interface{}) *AdvancedSelectBuilder {
	b.whereClause = append(b.whereClause, condition)
	if arg != nil {
		b.args = append(b.args, arg)
	}
	return b
}

// OrderBy adds sort orders.
func (b *AdvancedSelectBuilder) OrderBy(order ...string) *AdvancedSelectBuilder {
	b.orderBy = append(b.orderBy, order...)
	return b
}

// Limit sets query limit.
func (b *AdvancedSelectBuilder) Limit(limit int) *AdvancedSelectBuilder {
	b.limitVal = limit
	return b
}

// Build generates the complete SQL statement and its parameter list.
func (b *AdvancedSelectBuilder) Build() (string, []interface{}, error) {
	if b.table == "" && len(b.ctes) == 0 {
		return "", nil, fmt.Errorf("table or CTE required")
	}
	if len(b.projections) == 0 {
		b.projections = []string{"*"}
	}

	var sb strings.Builder

	if len(b.ctes) > 0 {
		sb.WriteString("WITH ")
		sb.WriteString(strings.Join(b.ctes, ", "))
		sb.WriteString(" ")
	}

	sb.WriteString("SELECT ")
	sb.WriteString(strings.Join(b.projections, ", "))

	if b.table != "" {
		sb.WriteString(" FROM ")
		sb.WriteString(b.table)
	}

	if len(b.whereClause) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(b.whereClause, " AND "))
	}

	if len(b.windows) > 0 {
		sb.WriteString(" WINDOW ")
		winDefs := make([]string, 0, len(b.windows))
		for name, spec := range b.windows {
			parts := []string{}
			if len(spec.PartitionBy) > 0 {
				parts = append(parts, "PARTITION BY "+strings.Join(spec.PartitionBy, ", "))
			}
			if len(spec.OrderBy) > 0 {
				parts = append(parts, "ORDER BY "+strings.Join(spec.OrderBy, ", "))
			}
			winDefs = append(winDefs, fmt.Sprintf("%s AS (%s)", name, strings.Join(parts, " ")))
		}
		sb.WriteString(strings.Join(winDefs, ", "))
	}

	if len(b.orderBy) > 0 {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(strings.Join(b.orderBy, ", "))
	}

	if b.limitVal > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", b.limitVal))
	}

	return sb.String(), b.args, nil
}
