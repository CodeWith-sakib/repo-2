package postgres

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// OpaqueTimeCursor stores pagination state for deterministic, index-supported queries.
type OpaqueTimeCursor struct {
	LastTimestamp time.Time `json:"last_timestamp"`
	LastID        string    `json:"last_id"`
}

// QueryCursorPaginator encodes and decodes forward/backward cursor tokens.
type QueryCursorPaginator struct {
	mu sync.RWMutex
}

// NewQueryCursorPaginator creates a cursor encoder.
func NewQueryCursorPaginator() *QueryCursorPaginator {
	return &QueryCursorPaginator{}
}

// EncodeCursor turns a timestamp and row ID into an opaque Base64 URL-safe token.
func (p *QueryCursorPaginator) EncodeCursor(ts time.Time, id string) string {
	raw := fmt.Sprintf("%d:%s", ts.UnixNano(), id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor decodes an opaque cursor token back into timestamp and row ID.
func (p *QueryCursorPaginator) DecodeCursor(token string) (*OpaqueTimeCursor, error) {
	bytes, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor encoding: %w", err)
	}

	parts := strings.SplitN(string(bytes), ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("malformed cursor payload")
	}

	nano, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor timestamp: %w", err)
	}

	return &OpaqueTimeCursor{
		LastTimestamp: time.Unix(0, nano),
		LastID:        parts[1],
	}, nil
}

// BuildWhereClause constructs standard SQL clause for keyset pagination.
func (p *QueryCursorPaginator) BuildWhereClause(cursor *OpaqueTimeCursor) string {
	if cursor == nil {
		return ""
	}
	return fmt.Sprintf("(created_at, id) < ('%s', '%s')",
		cursor.LastTimestamp.UTC().Format(time.RFC3339Nano), cursor.LastID)
}
