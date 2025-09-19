package postgres

import (
	"fmt"
	"sort"
	"strings"
)

// QueryHintType describes the type of optimizer hint.
type QueryHintType string

const (
	HintUseIndex       QueryHintType = "use_index"
	HintForceSeqScan   QueryHintType = "force_seqscan"
	HintParallelDegree QueryHintType = "parallel"
	HintHashJoin       QueryHintType = "hashjoin"
	HintMergeJoin      QueryHintType = "mergejoin"
	HintNestedLoop     QueryHintType = "nestloop"
)

// QueryHint is an optimizer directive embedded in query metadata.
type QueryHint struct {
	Type  QueryHintType
	Value string
}

// PlanJoinType describes how two tables are joined in a QueryPlan.
type PlanJoinType string

const (
	PlanJoinInner PlanJoinType = "INNER JOIN"
	PlanJoinLeft  PlanJoinType = "LEFT JOIN"
	PlanJoinRight PlanJoinType = "RIGHT JOIN"
	PlanJoinFull  PlanJoinType = "FULL OUTER JOIN"
	PlanJoinCross PlanJoinType = "CROSS JOIN"
)

// PlanJoinClause represents a JOIN clause in the query planner.
type PlanJoinClause struct {
	Type      PlanJoinType
	Table     string
	Alias     string
	Condition string
}

// ColumnExpression is a column or expression in a SELECT or WHERE clause.
type ColumnExpression struct {
	Expr  string
	Alias string
}

// OrderByClause describes a sort directive.
type OrderByClause struct {
	Expr       string
	Descending bool
	NullsLast  bool
}

// QueryPlan is the high-level logical representation of a planned SQL query.
type QueryPlan struct {
	BaseTable    string
	BaseAlias    string
	Columns      []ColumnExpression
	Joins        []PlanJoinClause
	WhereClause  []string
	GroupByCols  []string
	HavingClause string
	OrderBy      []OrderByClause
	Limit        int
	Offset       int
	Distinct     bool
	ForUpdate    bool
	Hints        []QueryHint
	CTEs         map[string]string
}

// QueryPlanner provides a builder interface for constructing complex PostgreSQL query plans.
type QueryPlanner struct {
	plan QueryPlan
}

// NewQueryPlanner initializes a planner for the given base table.
func NewQueryPlanner(table, alias string) *QueryPlanner {
	return &QueryPlanner{
		plan: QueryPlan{
			BaseTable: table,
			BaseAlias: alias,
			CTEs:      make(map[string]string),
		},
	}
}

// Select adds projected columns or expressions.
func (p *QueryPlanner) Select(cols ...ColumnExpression) *QueryPlanner {
	p.plan.Columns = append(p.plan.Columns, cols...)
	return p
}

// Join adds a JOIN clause.
func (p *QueryPlanner) Join(joinType PlanJoinType, table, alias, condition string) *QueryPlanner {
	p.plan.Joins = append(p.plan.Joins, PlanJoinClause{
		Type:      joinType,
		Table:     table,
		Alias:     alias,
		Condition: condition,
	})
	return p
}

// Where adds a WHERE predicate.
func (p *QueryPlanner) Where(pred string) *QueryPlanner {
	p.plan.WhereClause = append(p.plan.WhereClause, pred)
	return p
}

// GroupBy adds GROUP BY columns.
func (p *QueryPlanner) GroupBy(cols ...string) *QueryPlanner {
	p.plan.GroupByCols = append(p.plan.GroupByCols, cols...)
	return p
}

// Having sets the HAVING clause.
func (p *QueryPlanner) Having(cond string) *QueryPlanner {
	p.plan.HavingClause = cond
	return p
}

// OrderByAsc adds an ascending ORDER BY expression.
func (p *QueryPlanner) OrderByAsc(expr string) *QueryPlanner {
	p.plan.OrderBy = append(p.plan.OrderBy, OrderByClause{Expr: expr})
	return p
}

// OrderByDesc adds a descending ORDER BY expression.
func (p *QueryPlanner) OrderByDesc(expr string) *QueryPlanner {
	p.plan.OrderBy = append(p.plan.OrderBy, OrderByClause{Expr: expr, Descending: true})
	return p
}

// LimitOffset sets LIMIT and OFFSET for pagination.
func (p *QueryPlanner) LimitOffset(limit, offset int) *QueryPlanner {
	p.plan.Limit = limit
	p.plan.Offset = offset
	return p
}

// Distinct marks the query as DISTINCT.
func (p *QueryPlanner) Distinct() *QueryPlanner {
	p.plan.Distinct = true
	return p
}

// ForUpdate adds a FOR UPDATE locking clause.
func (p *QueryPlanner) ForUpdate() *QueryPlanner {
	p.plan.ForUpdate = true
	return p
}

// WithCTE adds a Common Table Expression.
func (p *QueryPlanner) WithCTE(name, query string) *QueryPlanner {
	p.plan.CTEs[name] = query
	return p
}

// Hint adds an optimizer hint.
func (p *QueryPlanner) Hint(h QueryHint) *QueryPlanner {
	p.plan.Hints = append(p.plan.Hints, h)
	return p
}

// Build generates the SQL string from the current plan.
func (p *QueryPlanner) Build() (string, error) {
	if p.plan.BaseTable == "" {
		return "", fmt.Errorf("query planner: base table is required")
	}

	var sb strings.Builder

	// CTEs
	if len(p.plan.CTEs) > 0 {
		names := make([]string, 0, len(p.plan.CTEs))
		for k := range p.plan.CTEs {
			names = append(names, k)
		}
		sort.Strings(names)

		sb.WriteString("WITH ")
		parts := make([]string, 0, len(names))
		for _, n := range names {
			parts = append(parts, fmt.Sprintf("%s AS (%s)", n, p.plan.CTEs[n]))
		}
		sb.WriteString(strings.Join(parts, ", "))
		sb.WriteString(" ")
	}

	// SELECT
	sb.WriteString("SELECT ")
	if p.plan.Distinct {
		sb.WriteString("DISTINCT ")
	}

	if len(p.plan.Columns) == 0 {
		base := p.plan.BaseAlias
		if base == "" {
			base = p.plan.BaseTable
		}
		sb.WriteString(base + ".*")
	} else {
		colParts := make([]string, len(p.plan.Columns))
		for i, c := range p.plan.Columns {
			if c.Alias != "" {
				colParts[i] = fmt.Sprintf("%s AS %s", c.Expr, c.Alias)
			} else {
				colParts[i] = c.Expr
			}
		}
		sb.WriteString(strings.Join(colParts, ", "))
	}

	// FROM
	sb.WriteString(" FROM ")
	sb.WriteString(p.plan.BaseTable)
	if p.plan.BaseAlias != "" {
		sb.WriteString(" " + p.plan.BaseAlias)
	}

	// JOINs
	for _, j := range p.plan.Joins {
		sb.WriteString(fmt.Sprintf(" %s %s %s ON %s", j.Type, j.Table, j.Alias, j.Condition))
	}

	// WHERE
	if len(p.plan.WhereClause) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(p.plan.WhereClause, " AND "))
	}

	// GROUP BY
	if len(p.plan.GroupByCols) > 0 {
		sb.WriteString(" GROUP BY ")
		sb.WriteString(strings.Join(p.plan.GroupByCols, ", "))
	}

	// HAVING
	if p.plan.HavingClause != "" {
		sb.WriteString(" HAVING ")
		sb.WriteString(p.plan.HavingClause)
	}

	// ORDER BY
	if len(p.plan.OrderBy) > 0 {
		sb.WriteString(" ORDER BY ")
		parts := make([]string, len(p.plan.OrderBy))
		for i, o := range p.plan.OrderBy {
			dir := ""
			if o.Descending {
				dir = " DESC"
			}
			nulls := ""
			if o.NullsLast {
				nulls = " NULLS LAST"
			}
			parts[i] = o.Expr + dir + nulls
		}
		sb.WriteString(strings.Join(parts, ", "))
	}

	// LIMIT / OFFSET
	if p.plan.Limit > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", p.plan.Limit))
	}
	if p.plan.Offset > 0 {
		sb.WriteString(fmt.Sprintf(" OFFSET %d", p.plan.Offset))
	}

	// FOR UPDATE
	if p.plan.ForUpdate {
		sb.WriteString(" FOR UPDATE")
	}

	return sb.String(), nil
}
