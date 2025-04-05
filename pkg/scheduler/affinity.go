package scheduler

import (
	"strings"
)

type WorkerCapability struct {
	WorkerID string            `json:"worker_id"`
	Labels   map[string]string `json:"labels"`
	BusySlots int              `json:"busy_slots"`
	MaxSlots  int              `json:"max_slots"`
}

type NodeAffinity struct {
	RequiredLabels map[string]string `json:"required_labels"`
	PreferredKeys  []string          `json:"preferred_keys"`
}

type AffinityMatcher struct{}

func NewAffinityMatcher() *AffinityMatcher {
	return &AffinityMatcher{}
}

// Matches returns true if worker satisfies all required labels.
func (m *AffinityMatcher) Matches(worker *WorkerCapability, affinity *NodeAffinity) bool {
	if affinity == nil {
		return true
	}
	for k, v := range affinity.RequiredLabels {
		workerVal, exists := worker.Labels[k]
		if !exists || !strings.EqualFold(workerVal, v) {
			return false
		}
	}
	return true
}

// Score ranks matched workers based on preferred keys and available capacity.
func (m *AffinityMatcher) Score(worker *WorkerCapability, affinity *NodeAffinity) int {
	score := (worker.MaxSlots - worker.BusySlots) * 10
	if affinity == nil {
		return score
	}
	for _, key := range affinity.PreferredKeys {
		if _, exists := worker.Labels[key]; exists {
			score += 25
		}
	}
	return score
}

// SelectBestWorker returns the highest scored worker matching affinity requirements.
func (m *AffinityMatcher) SelectBestWorker(workers []*WorkerCapability, affinity *NodeAffinity) *WorkerCapability {
	var best *WorkerCapability
	bestScore := -1

	for _, w := range workers {
		if w.BusySlots >= w.MaxSlots {
			continue
		}
		if !m.Matches(w, affinity) {
			continue
		}
		s := m.Score(w, affinity)
		if s > bestScore {
			bestScore = s
			best = w
		}
	}
	return best
}
