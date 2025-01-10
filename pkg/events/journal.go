package events

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"sync"
	"time"
)

type JournalRecord struct {
	Index     uint64
	Timestamp time.Time
	Payload   []byte
	Checksum  uint32
}

type MemoryJournal struct {
	mu      sync.RWMutex
	records []*JournalRecord
	seq     uint64
}

func NewMemoryJournal() *MemoryJournal {
	return &MemoryJournal{
		records: make([]*JournalRecord, 0),
	}
}

func (j *MemoryJournal) Append(payload []byte) (*JournalRecord, error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.seq++
	rec := &JournalRecord{
		Index:     j.seq,
		Timestamp: time.Now().UTC(),
		Payload:   append([]byte(nil), payload...),
		Checksum:  crc32.ChecksumIEEE(payload),
	}
	j.records = append(j.records, rec)
	return rec, nil
}

func (j *MemoryJournal) ReadFrom(fromIndex uint64, maxRecords int) ([]*JournalRecord, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	var result []*JournalRecord
	for _, r := range j.records {
		if r.Index >= fromIndex {
			// Verify checksum integrity
			expected := crc32.ChecksumIEEE(r.Payload)
			if r.Checksum != expected {
				return nil, fmt.Errorf("data corruption detected at index %d: checksum mismatch", r.Index)
			}
			result = append(result, r)
			if maxRecords > 0 && len(result) >= maxRecords {
				break
			}
		}
	}
	return result, nil
}

func EncodeRecord(w io.Writer, r *JournalRecord) error {
	buf := make([]byte, 8+8+4+4)
	binary.BigEndian.PutUint64(buf[0:8], r.Index)
	binary.BigEndian.PutUint64(buf[8:16], uint64(r.Timestamp.UnixNano()))
	binary.BigEndian.PutUint32(buf[16:20], uint32(len(r.Payload)))
	binary.BigEndian.PutUint32(buf[20:24], r.Checksum)

	if _, err := w.Write(buf); err != nil {
		return err
	}
	_, err := w.Write(r.Payload)
	return err
}
