package postgres

import (
	"testing"
)

func TestQueryReadWriteSplitter(t *testing.T) {
	_ = NewQueryReadWriteSplitter(DBClusterPool{})

	if !IsReadOnlyQuery("SELECT * FROM workflows WHERE tenant_id = $1") {
		t.Error("expected SELECT to be classified as read-only")
	}

	if IsReadOnlyQuery("SELECT * FROM workflows FOR UPDATE") {
		t.Error("SELECT FOR UPDATE must NOT be read-only")
	}

	if IsReadOnlyQuery("INSERT INTO workflows (id) VALUES ($1)") {
		t.Error("INSERT must NOT be read-only")
	}

	if IsReadOnlyQuery("UPDATE workflows SET status = $1") {
		t.Error("UPDATE must NOT be read-only")
	}
}
