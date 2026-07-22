package postgres

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// CachedPreparedStatement tracks metadata and execution counts of server-side prepared statements.
type CachedPreparedStatement struct {
	QueryHash      string    `json:"query_hash"`
	StatementName  string    `json:"statement_name"`
	SQLTemplate    string    `json:"sql_template"`
	ExecutionCount int64     `json:"execution_count"`
	LastUsedAt     time.Time `json:"last_used_at"`
}

// PreparedStatementCache manages a registry of prepared SQL statements per connection pool.
type PreparedStatementCache struct {
	mu         sync.RWMutex
	statements map[string]*CachedPreparedStatement
	maxCache   int
}

// NewPreparedStatementCache creates a cache for prepared queries.
func NewPreparedStatementCache(maxCached int) *PreparedStatementCache {
	if maxCached <= 0 {
		maxCached = 256
	}
	return &PreparedStatementCache{
		statements: make(map[string]*CachedPreparedStatement),
		maxCache:   maxCached,
	}
}

// GetOrPrepare returns existing prepared statement or prepares a unique statement name.
func (c *PreparedStatementCache) GetOrPrepare(sql string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hash := sha256.Sum256([]byte(sql))
	hashKey := hex.EncodeToString(hash[:8])

	if stmt, exists := c.statements[hashKey]; exists {
		stmt.ExecutionCount++
		stmt.LastUsedAt = time.Now()
		return stmt.StatementName, true
	}

	stmtName := fmt.Sprintf("kf_stmt_%s", hashKey)
	c.statements[hashKey] = &CachedPreparedStatement{
		QueryHash:      hashKey,
		StatementName:  stmtName,
		SQLTemplate:    sql,
		ExecutionCount: 1,
		LastUsedAt:     time.Now(),
	}

	return stmtName, false
}

// StatementCount returns number of active prepared statements.
func (c *PreparedStatementCache) StatementCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.statements)
}
