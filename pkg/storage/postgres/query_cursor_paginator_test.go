package postgres

import (
	"testing"
	"time"
)

func TestQueryCursorPaginator(t *testing.T) {
	paginator := NewQueryCursorPaginator()
	now := time.Date(2026, 6, 6, 12, 0, 0, 123456789, time.UTC)
	id := "run_abcdef123"

	token := paginator.EncodeCursor(now, id)
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	cursor, err := paginator.DecodeCursor(token)
	if err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
	if !cursor.LastTimestamp.Equal(now) {
		t.Errorf("timestamp mismatch: got %v vs %v", cursor.LastTimestamp, now)
	}
	if cursor.LastID != id {
		t.Errorf("id mismatch: got %s vs %s", cursor.LastID, id)
	}

	sql := paginator.BuildWhereClause(cursor)
	if sql == "" {
		t.Error("expected valid WHERE clause")
	}
}
