package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

type QueryExplainer struct {
	db *sql.DB
}

func NewQueryExplainer(db *sql.DB) *QueryExplainer {
	return &QueryExplainer{db: db}
}

func (e *QueryExplainer) Explain(ctx context.Context, query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("empty query")
	}
	return "EXPLAIN " + query, nil
}
