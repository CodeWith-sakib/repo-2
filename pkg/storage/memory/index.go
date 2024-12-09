package memory

import (
	"sync"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type SecondaryIndex struct {
	mu    sync.RWMutex
	index map[string]map[core.ID]bool
}

func NewSecondaryIndex() *SecondaryIndex {
	return &SecondaryIndex{
		index: make(map[string]map[core.ID]bool),
	}
}

func (idx *SecondaryIndex) Index(key string, id core.ID) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if _, exists := idx.index[key]; !exists {
		idx.index[key] = make(map[core.ID]bool)
	}
	idx.index[key][id] = true
}

func (idx *SecondaryIndex) Unindex(key string, id core.ID) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if m, exists := idx.index[key]; exists {
		delete(m, id)
		if len(m) == 0 {
			delete(idx.index, key)
		}
	}
}

func (idx *SecondaryIndex) Lookup(key string) []core.ID {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	m, exists := idx.index[key]
	if !exists {
		return nil
	}

	result := make([]core.ID, 0, len(m))
	for id := range m {
		result = append(result, id)
	}
	return result
}
