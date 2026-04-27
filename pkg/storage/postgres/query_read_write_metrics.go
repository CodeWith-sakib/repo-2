package postgres

import (
	"sync/atomic"
	"time"
)

// ReadWriteTrafficMetrics records separate execution counters for read replicas and primary master.
type ReadWriteTrafficMetrics struct {
	MasterQueries  atomic.Int64
	ReplicaQueries atomic.Int64
	TotalErrors    atomic.Int64
	LastRecorded   atomic.Int64 // unix nanos
}

// NewReadWriteTrafficMetrics creates a traffic counter.
func NewReadWriteTrafficMetrics() *ReadWriteTrafficMetrics {
	m := &ReadWriteTrafficMetrics{}
	m.LastRecorded.Store(time.Now().UTC().UnixNano())
	return m
}

// RecordMasterQuery increments master write query counter.
func (m *ReadWriteTrafficMetrics) RecordMasterQuery() {
	m.MasterQueries.Add(1)
	m.LastRecorded.Store(time.Now().UTC().UnixNano())
}

// RecordReplicaQuery increments replica read query counter.
func (m *ReadWriteTrafficMetrics) RecordReplicaQuery() {
	m.ReplicaQueries.Add(1)
	m.LastRecorded.Store(time.Now().UTC().UnixNano())
}

// RecordError increments query error tally.
func (m *ReadWriteTrafficMetrics) RecordError() {
	m.TotalErrors.Add(1)
}

// ReadRatioPct returns percentage of queries served by read replicas.
func (m *ReadWriteTrafficMetrics) ReadRatioPct() float64 {
	master := m.MasterQueries.Load()
	replica := m.ReplicaQueries.Load()
	total := master + replica
	if total == 0 {
		return 0.0
	}
	return (float64(replica) / float64(total)) * 100.0
}
