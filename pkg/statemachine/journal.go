package statemachine

import (
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type TransitionRecord struct {
	RunID     core.ID
	FromState core.RunState
	ToState   core.RunState
	Timestamp time.Time
	Reason    string
}

type StateTransitionJournal struct {
	mu      sync.RWMutex
	records []*TransitionRecord
}

func NewStateTransitionJournal() *StateTransitionJournal {
	return &StateTransitionJournal{
		records: make([]*TransitionRecord, 0),
	}
}

func (j *StateTransitionJournal) RecordTransition(runID core.ID, from, to core.RunState, reason string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.records = append(j.records, &TransitionRecord{
		RunID:     runID,
		FromState: from,
		ToState:   to,
		Timestamp: time.Now().UTC(),
		Reason:    reason,
	})
}

func (j *StateTransitionJournal) HistoryForRun(runID core.ID) []*TransitionRecord {
	j.mu.RLock()
	defer j.mu.RUnlock()
	var out []*TransitionRecord
	for _, rec := range j.records {
		if rec.RunID == runID {
			out = append(out, rec)
		}
	}
	return out
}
