package storage

import (
	"testing"
)

func TestWALManagerAppendAndReplay(t *testing.T) {
	wal := NewWALManager(2)

	r1, err := wal.Append(WALRecordRunStateChange, []byte("run:running"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r2, err := wal.Append(WALRecordStepStateChange, []byte("step:completed"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r3, err := wal.Append(WALRecordCheckpoint, []byte("checkpoint:1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r1.LSN != 1 || r2.LSN != 2 || r3.LSN != 3 {
		t.Errorf("unexpected LSN progression: %d, %d, %d", r1.LSN, r2.LSN, r3.LSN)
	}

	replayed := 0
	err = wal.ReplayFrom(2, func(rec *WALRecord) error {
		replayed++
		return nil
	})
	if err != nil {
		t.Fatalf("replay failed: %v", err)
	}
	if replayed != 2 {
		t.Errorf("expected 2 records replayed from LSN 2, got %d", replayed)
	}
}
