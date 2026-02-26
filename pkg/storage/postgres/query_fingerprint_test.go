package postgres

import (
	"testing"
)

func TestQueryFingerprinter(t *testing.T) {
	fp := NewQueryFingerprinter()

	q1 := "SELECT * FROM workflow_runs WHERE tenant_id = 'acme' AND retry_count > 3"
	q2 := "select *   from  workflow_runs  where tenant_id = 'globex'  and retry_count > 10"

	norm1 := fp.Normalize(q1)
	norm2 := fp.Normalize(q2)

	if norm1 != norm2 {
		t.Errorf("expected identical normalized queries: %s vs %s", norm1, norm2)
	}

	f1 := fp.Fingerprint(q1)
	f2 := fp.Fingerprint(q2)

	if f1 != f2 {
		t.Errorf("expected identical fingerprints: %s vs %s", f1, f2)
	}

	if len(f1) != 16 {
		t.Errorf("expected 16 hex char fingerprint, got %d chars: %s", len(f1), f1)
	}
}
