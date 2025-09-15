package maintenance

import (
	"testing"
	"time"
)

func TestCompactionScheduler_DrainAndResult(t *testing.T) {
	sched := NewCompactionScheduler(4)

	sched.RegisterRunner("events", func(job *CompactionJob) CompactionResult {
		time.Sleep(5 * time.Millisecond)
		return CompactionResult{
			JobID:       job.ID,
			BytesBefore: 10000,
			BytesAfter:  4000,
		}
	})

	for i := 0; i < 5; i++ {
		_, err := sched.Enqueue("events", i, 10-i)
		if err != nil {
			t.Fatalf("enqueue job %d: %v", i, err)
		}
	}

	if sched.PendingCount() != 5 {
		t.Errorf("expected 5 pending, got %d", sched.PendingCount())
	}

	completed := sched.Drain()

	if len(completed) != 5 {
		t.Errorf("expected 5 completed, got %d", len(completed))
	}

	for _, j := range completed {
		if j.Status != CompactionCompleted {
			t.Errorf("job %s expected completed, got %s: %s", j.ID, j.Status, j.ErrMsg)
		}
		if j.BytesAfter >= j.BytesBefore {
			t.Errorf("job %s: expected compression, got before=%d after=%d", j.ID, j.BytesBefore, j.BytesAfter)
		}
	}

	if sched.PendingCount() != 0 {
		t.Errorf("expected 0 pending after drain, got %d", sched.PendingCount())
	}
}

func TestCompactionScheduler_UnknownTable(t *testing.T) {
	sched := NewCompactionScheduler(2)
	if _, err := sched.Enqueue("nonexistent", 0, 1); err == nil {
		t.Error("expected error for unknown table")
	}
}
