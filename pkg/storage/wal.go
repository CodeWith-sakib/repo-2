package storage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"sync"
	"time"
)

type WALRecordType uint8

const (
	WALRecordRunStateChange WALRecordType = 1
	WALRecordStepStateChange WALRecordType = 2
	WALRecordCheckpoint      WALRecordType = 3
)

type WALRecord struct {
	LSN       uint64
	Timestamp time.Time
	Type      WALRecordType
	Data      []byte
	Checksum  uint32
}

type WALSegment struct {
	SegmentID uint32
	records   []*WALRecord
}

type WALManager struct {
	mu           sync.RWMutex
	currentLSN   uint64
	segmentSize  int
	segments     []*WALSegment
	activeSegment *WALSegment
}

func NewWALManager(segmentSize int) *WALManager {
	if segmentSize <= 0 {
		segmentSize = 1000
	}
	initialSeg := &WALSegment{
		SegmentID: 1,
		records:   make([]*WALRecord, 0),
	}
	return &WALManager{
		segmentSize:   segmentSize,
		segments:      []*WALSegment{initialSeg},
		activeSegment: initialSeg,
	}
}

func (w *WALManager) Append(recType WALRecordType, data []byte) (*WALRecord, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.currentLSN++
	rec := &WALRecord{
		LSN:       w.currentLSN,
		Timestamp: time.Now().UTC(),
		Type:      recType,
		Data:      append([]byte(nil), data...),
		Checksum:  crc32.ChecksumIEEE(data),
	}

	if len(w.activeSegment.records) >= w.segmentSize {
		newSeg := &WALSegment{
			SegmentID: uint32(len(w.segments) + 1),
			records:   make([]*WALRecord, 0),
		}
		w.segments = append(w.segments, newSeg)
		w.activeSegment = newSeg
	}

	w.activeSegment.records = append(w.activeSegment.records, rec)
	return rec, nil
}

func (w *WALManager) ReplayFrom(startLSN uint64, handler func(rec *WALRecord) error) error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for _, seg := range w.segments {
		for _, rec := range seg.records {
			if rec.LSN >= startLSN {
				// Verify checksum
				if crc32.ChecksumIEEE(rec.Data) != rec.Checksum {
					return fmt.Errorf("WAL corruption detected at LSN %d", rec.LSN)
				}
				if err := handler(rec); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *WALRecord) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, r.LSN); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, r.Timestamp.UnixNano()); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, uint8(r.Type)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, uint32(len(r.Data))); err != nil {
		return nil, err
	}
	if _, err := buf.Write(r.Data); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, r.Checksum); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
