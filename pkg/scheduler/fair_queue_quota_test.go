package scheduler

import (
	"testing"
)

func TestQuotaEnforcer(t *testing.T) {
	qe := NewQuotaEnforcer()
	qe.SetQuota("tenant-alpha", TenantQuota{
		MaxConcurrentRuns: 1,
		MaxQueuedRuns:     2,
	})

	if !qe.Enqueue("tenant-alpha") {
		t.Error("expected first enqueue to succeed")
	}
	if !qe.Enqueue("tenant-alpha") {
		t.Error("expected second enqueue to succeed")
	}
	if qe.Enqueue("tenant-alpha") {
		t.Error("expected third enqueue to exceed queue quota")
	}

	if !qe.Start("tenant-alpha") {
		t.Error("expected first run to start")
	}
	if qe.Start("tenant-alpha") {
		t.Error("expected second run to exceed concurrent quota")
	}

	qe.Finish("tenant-alpha")
	if !qe.Start("tenant-alpha") {
		t.Error("expected run to start after previous finished")
	}
}
