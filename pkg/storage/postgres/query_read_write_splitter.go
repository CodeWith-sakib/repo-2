package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// DBClusterPool holds separated connections for write master and read replicas.
type DBClusterPool struct {
	Master   *sql.DB
	Replicas []*sql.DB
	rrIndex  int
}

// QueryReadWriteSplitter automatically routes read-only SELECT statements to read replicas.
type QueryReadWriteSplitter struct {
	cluster DBClusterPool
}

// NewQueryReadWriteSplitter creates an automatic read/write query router.
func NewQueryReadWriteSplitter(cluster DBClusterPool) *QueryReadWriteSplitter {
	return &QueryReadWriteSplitter{
		cluster: cluster,
	}
}

// IsReadOnlyQuery determines whether a query statement is safe for read replica execution.
func IsReadOnlyQuery(sqlQuery string) bool {
	q := strings.TrimSpace(strings.ToUpper(sqlQuery))
	if strings.HasPrefix(q, "SELECT") || strings.HasPrefix(q, "EXPLAIN") {
		// Ensure it's not a SELECT ... FOR UPDATE / SHARE
		if !strings.Contains(q, "FOR UPDATE") && !strings.Contains(q, "FOR SHARE") {
			return true
		}
	}
	return false
}

// RouteDB returns the appropriate database target (master or round-robin replica).
func (s *QueryReadWriteSplitter) RouteDB(sqlQuery string) (*sql.DB, bool) {
	if IsReadOnlyQuery(sqlQuery) && len(s.cluster.Replicas) > 0 {
		idx := s.cluster.rrIndex % len(s.cluster.Replicas)
		s.cluster.rrIndex++
		return s.cluster.Replicas[idx], true
	}
	return s.cluster.Master, false
}

// ExecContext routes writes to master.
func (s *QueryReadWriteSplitter) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	db, isReplica := s.RouteDB(query)
	if isReplica {
		return nil, errors.New("cannot execute mutating command on read replica")
	}
	if db == nil {
		return nil, nil // mock safe
	}
	return db.ExecContext(ctx, query, args...)
}
