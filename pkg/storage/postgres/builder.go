package postgres

import (
	"fmt"
	"strconv"
	"strings"
)

type SQLDialect int

const (
	DialectPostgreSQL SQLDialect = iota
	DialectSQLite
	DialectMySQL
)

type ClauseType string

const (
	ClauseSelect  ClauseType = "SELECT"
	ClauseFrom    ClauseType = "FROM"
	ClauseJoin    ClauseType = "JOIN"
	ClauseWhere   ClauseType = "WHERE"
	ClauseGroupBy ClauseType = "GROUP BY"
	ClauseHaving  ClauseType = "HAVING"
	ClauseOrderBy ClauseType = "ORDER BY"
	ClauseLimit   ClauseType = "LIMIT"
	ClauseOffset  ClauseType = "OFFSET"
)

type JoinType string

const (
	JoinInner JoinType = "INNER JOIN"
	JoinLeft  JoinType = "LEFT OUTER JOIN"
	JoinRight JoinType = "RIGHT OUTER JOIN"
	JoinFull  JoinType = "FULL OUTER JOIN"
)

type JoinClause struct {
	Type      JoinType
	Table     string
	Condition string
}

type OrderDirection string

const (
	OrderAsc  OrderDirection = "ASC"
	OrderDesc OrderDirection = "DESC"
)

type OrderItem struct {
	Column    string
	Direction OrderDirection
}

type QueryBuilder struct {
	dialect    SQLDialect
	table      string
	columns    []string
	joins      []JoinClause
	whereClauses []string
	groupBy    []string
	having     []string
	orderBy    []OrderItem
	limit      int
	offset     int
	args       []interface{}
	forUpdate  bool
	skipLocked bool
}

func NewQueryBuilder(dialect SQLDialect) *QueryBuilder {
	return &QueryBuilder{
		dialect:      dialect,
		columns:      make([]string, 0),
		joins:        make([]JoinClause, 0),
		whereClauses: make([]string, 0),
		groupBy:      make([]string, 0),
		having:       make([]string, 0),
		orderBy:      make([]OrderItem, 0),
		args:         make([]interface{}, 0),
	}
}

func (b *QueryBuilder) Table(table string) *QueryBuilder {
	b.table = table
	return b
}

func (b *QueryBuilder) Select(cols ...string) *QueryBuilder {
	b.columns = append(b.columns, cols...)
	return b
}

func (b *QueryBuilder) Join(jt JoinType, table, cond string) *QueryBuilder {
	b.joins = append(b.joins, JoinClause{Type: jt, Table: table, Condition: cond})
	return b
}

func (b *QueryBuilder) Where(cond string, args ...interface{}) *QueryBuilder {
	b.whereClauses = append(b.whereClauses, cond)
	b.args = append(b.args, args...)
	return b
}

func (b *QueryBuilder) And(cond string, args ...interface{}) *QueryBuilder {
	return b.Where(cond, args...)
}

func (b *QueryBuilder) GroupBy(cols ...string) *QueryBuilder {
	b.groupBy = append(b.groupBy, cols...)
	return b
}

func (b *QueryBuilder) OrderBy(col string, dir OrderDirection) *QueryBuilder {
	b.orderBy = append(b.orderBy, OrderItem{Column: col, Direction: dir})
	return b
}

func (b *QueryBuilder) Limit(limit int) *QueryBuilder {
	b.limit = limit
	return b
}

func (b *QueryBuilder) Offset(offset int) *QueryBuilder {
	b.offset = offset
	return b
}

func (b *QueryBuilder) ForUpdate(skipLocked bool) *QueryBuilder {
	b.forUpdate = true
	b.skipLocked = skipLocked
	return b
}

func (b *QueryBuilder) Build() (string, []interface{}) {
	var sb strings.Builder
	sb.WriteString("SELECT ")

	if len(b.columns) == 0 {
		sb.WriteString("*")
	} else {
		sb.WriteString(strings.Join(b.columns, ", "))
	}

	sb.WriteString(" FROM ")
	sb.WriteString(b.table)

	for _, j := range b.joins {
		sb.WriteString(" ")
		sb.WriteString(string(j.Type))
		sb.WriteString(" ")
		sb.WriteString(j.Table)
		sb.WriteString(" ON ")
		sb.WriteString(j.Condition)
	}

	paramIndex := 1
	if len(b.whereClauses) > 0 {
		sb.WriteString(" WHERE ")
		for i, w := range b.whereClauses {
			if i > 0 {
				sb.WriteString(" AND ")
			}
			rebound, nextIdx := rebindPlaceholders(w, paramIndex, b.dialect)
			sb.WriteString(rebound)
			paramIndex = nextIdx
		}
	}

	if len(b.groupBy) > 0 {
		sb.WriteString(" GROUP BY ")
		sb.WriteString(strings.Join(b.groupBy, ", "))
	}

	if len(b.orderBy) > 0 {
		sb.WriteString(" ORDER BY ")
		items := make([]string, len(b.orderBy))
		for i, o := range b.orderBy {
			items[i] = fmt.Sprintf("%s %s", o.Column, o.Direction)
		}
		sb.WriteString(strings.Join(items, ", "))
	}

	if b.limit > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", b.limit))
	}
	if b.offset > 0 {
		sb.WriteString(fmt.Sprintf(" OFFSET %d", b.offset))
	}

	if b.forUpdate {
		sb.WriteString(" FOR UPDATE")
		if b.skipLocked {
			sb.WriteString(" SKIP LOCKED")
		}
	}

	return sb.String(), b.args
}

type InsertBuilder struct {
	dialect SQLDialect
	table   string
	columns []string
	values  [][]interface{}
}

func NewInsertBuilder(dialect SQLDialect) *InsertBuilder {
	return &InsertBuilder{
		dialect: dialect,
		columns: make([]string, 0),
		values:  make([][]interface{}, 0),
	}
}

func (ib *InsertBuilder) Into(table string) *InsertBuilder {
	ib.table = table
	return ib
}

func (ib *InsertBuilder) Columns(cols ...string) *InsertBuilder {
	ib.columns = append(ib.columns, cols...)
	return ib
}

func (ib *InsertBuilder) Values(vals ...interface{}) *InsertBuilder {
	ib.values = append(ib.values, vals)
	return ib
}

func (ib *InsertBuilder) Build() (string, []interface{}) {
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	sb.WriteString(ib.table)
	sb.WriteString(" (")
	sb.WriteString(strings.Join(ib.columns, ", "))
	sb.WriteString(") VALUES ")

	var allArgs []interface{}
	paramIdx := 1

	for rowIdx, row := range ib.values {
		if rowIdx > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("(")
		for colIdx, val := range row {
			if colIdx > 0 {
				sb.WriteString(", ")
			}
			if ib.dialect == DialectPostgreSQL {
				sb.WriteString("$" + strconv.Itoa(paramIdx))
			} else {
				sb.WriteString("?")
			}
			paramIdx++
			allArgs = append(allArgs, val)
		}
		sb.WriteString(")")
	}

	return sb.String(), allArgs
}

func rebindPlaceholders(clause string, startIdx int, dialect SQLDialect) (string, int) {
	if dialect != DialectPostgreSQL {
		return clause, startIdx
	}

	var sb strings.Builder
	idx := startIdx
	for i := 0; i < len(clause); i++ {
		if clause[i] == '?' {
			sb.WriteString("$" + strconv.Itoa(idx))
			idx++
		} else {
			sb.WriteByte(clause[i])
		}
	}
	return sb.String(), idx
}
