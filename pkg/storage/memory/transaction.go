package memory

import (
	"fmt"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type VersionedRecord struct {
	Data      interface{}
	Version   int64
	Timestamp time.Time
	Deleted   bool
}

type MVCCTransaction struct {
	txID      int64
	store     *MVCCStore
	readView  int64
	writes    map[string]interface{}
	deletes   map[string]bool
	committed bool
	aborted   bool
	mu        sync.Mutex
}

type MVCCStore struct {
	mu           sync.RWMutex
	globalTxID   int64
	records      map[string][]*VersionedRecord
	activeTxs    map[int64]*MVCCTransaction
}

func NewMVCCStore() *MVCCStore {
	return &MVCCStore{
		records:   make(map[string][]*VersionedRecord),
		activeTxs: make(map[int64]*MVCCTransaction),
	}
}

func (s *MVCCStore) Begin() *MVCCTransaction {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.globalTxID++
	tx := &MVCCTransaction{
		txID:     s.globalTxID,
		store:    s,
		readView: s.globalTxID,
		writes:   make(map[string]interface{}),
		deletes:  make(map[string]bool),
	}
	s.activeTxs[tx.txID] = tx
	return tx
}

func (tx *MVCCTransaction) Put(key string, val interface{}) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.committed || tx.aborted {
		return fmt.Errorf("transaction already finished")
	}

	delete(tx.deletes, key)
	tx.writes[key] = val
	return nil
}

func (tx *MVCCTransaction) Get(key string) (interface{}, bool, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.committed || tx.aborted {
		return nil, false, fmt.Errorf("transaction already finished")
	}

	if tx.deletes[key] {
		return nil, false, nil
	}
	if val, exists := tx.writes[key]; exists {
		return val, true, nil
	}

	// Read from store using snapshot isolation (readView)
	tx.store.mu.RLock()
	defer tx.store.mu.RUnlock()

	versions, exists := tx.store.records[key]
	if !exists || len(versions) == 0 {
		return nil, false, nil
	}

	// Find newest version visible to readView
	for i := len(versions) - 1; i >= 0; i-- {
		ver := versions[i]
		if ver.Version <= tx.readView {
			if ver.Deleted {
				return nil, false, nil
			}
			return ver.Data, true, nil
		}
	}

	return nil, false, nil
}

func (tx *MVCCTransaction) Delete(key string) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.committed || tx.aborted {
		return fmt.Errorf("transaction already finished")
	}

	delete(tx.writes, key)
	tx.deletes[key] = true
	return nil
}

func (tx *MVCCTransaction) Commit() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.committed || tx.aborted {
		return fmt.Errorf("transaction already finished")
	}

	s := tx.store
	s.mu.Lock()
	defer s.mu.Unlock()

	// Conflict detection (First-Committer-Wins)
	for key := range tx.writes {
		versions := s.records[key]
		if len(versions) > 0 {
			latest := versions[len(versions)-1]
			if latest.Version > tx.readView {
				tx.aborted = true
				delete(s.activeTxs, tx.txID)
				return fmt.Errorf("%w: write conflict on key %s", core.ErrConflict, key)
			}
		}
	}

	commitTime := time.Now().UTC()
	s.globalTxID++
	commitVer := s.globalTxID

	for k, v := range tx.writes {
		rec := &VersionedRecord{
			Data:      v,
			Version:   commitVer,
			Timestamp: commitTime,
		}
		s.records[k] = append(s.records[k], rec)
	}

	for k := range tx.deletes {
		rec := &VersionedRecord{
			Version:   commitVer,
			Timestamp: commitTime,
			Deleted:   true,
		}
		s.records[k] = append(s.records[k], rec)
	}

	tx.committed = true
	delete(s.activeTxs, tx.txID)
	return nil
}

func (tx *MVCCTransaction) Rollback() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.committed {
		return fmt.Errorf("cannot rollback already committed transaction")
	}
	if tx.aborted {
		return nil
	}

	tx.aborted = true
	tx.store.mu.Lock()
	delete(tx.store.activeTxs, tx.txID)
	tx.store.mu.Unlock()
	return nil
}
