package storage

import "fmt"

import (
	"testing"
	"time"
)

func TestWALCompactor_CompactsAgedSegments(t *testing.T) {
	policy := WALCompactorPolicy{
		MaxSegmentAge:     50 * time.Millisecond,
		MaxSegmentBytes:   10 * 1024 * 1024, // 10MB
		MinSegmentsToKeep: 1,
	}

	comp := NewWALCompactor(policy)

	// 3 old segments
	for i := 0; i < 3; i++ {
		comp.RegisterSegment(WALSegmentMetadata{
			SegmentID:   fmt.Sprintf("seg-%03d", i),
			StartOffset: uint64(i) * 1024,
			EndOffset:   uint64(i)*1024 + 512,
			EntryCount:  100,
			CreatedAt:   time.Now().Add(-100 * time.Millisecond), // definitely expired
		})
	}

	// 1 fresh segment
	comp.RegisterSegment(WALSegmentMetadata{
		SegmentID:   "seg-fresh",
		StartOffset: 4096,
		EndOffset:   5000,
		EntryCount:  10,
		CreatedAt:   time.Now(),
	})

	result, err := comp.Compact()
	if err != nil {
		t.Fatalf("compact failed: %v", err)
	}

	if result.SegmentsScanned != 4 {
		t.Errorf("expected 4 scanned, got %d", result.SegmentsScanned)
	}
	// Should compact 3 old segments, keep at least 1 (minSegmentsToKeep)
	if result.SegmentsCompacted < 2 {
		t.Errorf("expected at least 2 compacted, got %d", result.SegmentsCompacted)
	}
	if result.BytesReclaimed == 0 {
		t.Error("expected non-zero bytes reclaimed")
	}
	if comp.PendingCount() < 1 {
		t.Error("expected at least 1 pending segment (minKeep)")
	}
}
