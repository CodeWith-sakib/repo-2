package postgres

import (
	"strings"
	"testing"
)

func TestQueryPlanner_BasicSelect(t *testing.T) {
	q, err := NewQueryPlanner("users", "u").
		Select(ColumnExpression{Expr: "u.id"}, ColumnExpression{Expr: "u.email"}).
		Where("u.active = true").
		OrderByAsc("u.created_at").
		LimitOffset(50, 0).
		Build()

	if err != nil {
		t.Fatalf("build error: %v", err)
	}
	if !strings.Contains(q, "SELECT u.id, u.email FROM users u") {
		t.Errorf("unexpected query: %s", q)
	}
	if !strings.Contains(q, "WHERE u.active = true") {
		t.Errorf("expected WHERE clause: %s", q)
	}
	if !strings.Contains(q, "LIMIT 50") {
		t.Errorf("expected LIMIT: %s", q)
	}
}

func TestQueryPlanner_WithJoinAndGroupBy(t *testing.T) {
	q, err := NewQueryPlanner("orders", "o").
		Select(ColumnExpression{Expr: "u.id", Alias: "user_id"}, ColumnExpression{Expr: "COUNT(*)", Alias: "order_count"}).
		Join(PlanJoinInner, "users", "u", "u.id = o.user_id").
		Where("o.status = 'completed'").
		GroupBy("u.id").
		Having("COUNT(*) > 5").
		OrderByDesc("order_count").
		Build()

	if err != nil {
		t.Fatalf("build error: %v", err)
	}
	if !strings.Contains(q, "INNER JOIN users u ON u.id = o.user_id") {
		t.Errorf("missing JOIN: %s", q)
	}
	if !strings.Contains(q, "GROUP BY u.id") {
		t.Errorf("missing GROUP BY: %s", q)
	}
	if !strings.Contains(q, "HAVING COUNT(*) > 5") {
		t.Errorf("missing HAVING: %s", q)
	}
	if !strings.Contains(q, "ORDER BY order_count DESC") {
		t.Errorf("missing DESC: %s", q)
	}
}

func TestQueryPlanner_WithCTE(t *testing.T) {
	q, err := NewQueryPlanner("ranked", "r").
		WithCTE("ranked", "SELECT *, ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at DESC) AS rn FROM events").
		Where("r.rn = 1").
		Build()

	if err != nil {
		t.Fatalf("build error: %v", err)
	}
	if !strings.Contains(q, "WITH ranked AS") {
		t.Errorf("missing CTE: %s", q)
	}
}

func TestQueryPlanner_NoTable(t *testing.T) {
	_, err := NewQueryPlanner("", "").Build()
	if err == nil {
		t.Error("expected error when no base table is set")
	}
}
