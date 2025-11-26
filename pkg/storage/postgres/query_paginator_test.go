package postgres

import (
	"strings"
	"testing"
)

func TestKeysetPaginationBuilder_FirstPage(t *testing.T) {
	builder, err := NewKeysetPaginationBuilder("workflows", "created_at", "id", true, 25)
	if err != nil {
		t.Fatalf("create builder failed: %v", err)
	}

	// First page (no cursor)
	sql, args := builder.BuildQuery(nil, "tenant_id = 'acme'", 1)

	if len(args) != 0 {
		t.Errorf("expected 0 args for first page, got %d", len(args))
	}
	if !strings.Contains(sql, "WHERE tenant_id = 'acme'") {
		t.Errorf("missing WHERE: %s", sql)
	}
	if !strings.Contains(sql, "ORDER BY created_at DESC, id DESC LIMIT 26") {
		t.Errorf("missing ORDER BY/LIMIT: %s", sql)
	}
}

func TestKeysetPaginationBuilder_WithCursor(t *testing.T) {
	builder, _ := NewKeysetPaginationBuilder("events", "timestamp", "id", true, 10)

	cursorStr, err := EncodeCursor("2026-03-01T12:00:00Z", "evt-99", SeekForward)
	if err != nil {
		t.Fatalf("encode cursor failed: %v", err)
	}

	cursor, err := DecodeCursor(cursorStr)
	if err != nil {
		t.Fatalf("decode cursor failed: %v", err)
	}

	sql, args := builder.BuildQuery(cursor, "", 1)

	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if !strings.Contains(sql, "WHERE (timestamp, id) < ($1, $2)") {
		t.Errorf("unexpected composite seek clause: %s", sql)
	}
	if args[0] != "2026-03-01T12:00:00Z" || args[1] != "evt-99" {
		t.Errorf("unexpected args: %+v", args)
	}
}
