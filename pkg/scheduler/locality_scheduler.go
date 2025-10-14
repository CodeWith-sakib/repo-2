package scheduler

import (
	"fmt"
	"sort"
	"sync"
)

// NodeTopology describes physical and network placement of a compute worker node.
type NodeTopology struct {
	NodeID      string
	Host        string
	Rack        string
	Zone        string
	Region      string
	ActiveTasks int
	Capacity    int
}

// LocalityPreference indicates where data dependencies reside for a task.
type LocalityPreference struct {
	PreferredHost   string
	PreferredRack   string
	PreferredZone   string
	PreferredRegion string
}

// ScoredNode pairs a node with its computed locality and balance score.
type ScoredNode struct {
	NodeID string
	Score  int
}

// LocalityScheduler selects optimal worker nodes based on data proximity and active workload.
type LocalityScheduler struct {
	mu    sync.RWMutex
	nodes map[string]*NodeTopology
}

// NewLocalityScheduler creates a new locality scheduler instance.
func NewLocalityScheduler() *LocalityScheduler {
	return &LocalityScheduler{
		nodes: make(map[string]*NodeTopology),
	}
}

// RegisterNode adds or updates a node's topology and capacity in the scheduler.
func (s *LocalityScheduler) RegisterNode(node NodeTopology) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodes[node.NodeID] = &node
}

// UnregisterNode removes a node.
func (s *LocalityScheduler) UnregisterNode(nodeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.nodes, nodeID)
}

// SelectNode finds the highest-scoring candidate node for a task with given locality preferences.
func (s *LocalityScheduler) SelectNode(pref LocalityPreference) (*NodeTopology, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.nodes) == 0 {
		return nil, fmt.Errorf("no worker nodes available")
	}

	var candidates []ScoredNode

	for _, node := range s.nodes {
		if node.Capacity > 0 && node.ActiveTasks >= node.Capacity {
			// Saturated node, skip
			continue
		}

		score := s.computeScore(node, pref)
		candidates = append(candidates, ScoredNode{
			NodeID: node.NodeID,
			Score:  score,
		})
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("all worker nodes are at maximum capacity")
	}

	// Sort candidates by score descending, then by NodeID ascending for determinism
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].NodeID < candidates[j].NodeID
		}
		return candidates[i].Score > candidates[j].Score
	})

	best := s.nodes[candidates[0].NodeID]
	return best, nil
}

func (s *LocalityScheduler) computeScore(node *NodeTopology, pref LocalityPreference) int {
	score := 0

	// Proximity points
	if pref.PreferredHost != "" && node.Host == pref.PreferredHost {
		score += 100 // Host-local match!
	} else if pref.PreferredRack != "" && node.Rack == pref.PreferredRack {
		score += 60 // Rack-local match
	} else if pref.PreferredZone != "" && node.Zone == pref.PreferredZone {
		score += 30 // Zone-local match
	} else if pref.PreferredRegion != "" && node.Region == pref.PreferredRegion {
		score += 10 // Region-local match
	}

	// Load penalty: deduct points for active tasks to avoid hotspotting
	if node.Capacity > 0 {
		loadPct := (float64(node.ActiveTasks) / float64(node.Capacity)) * 25.0
		score -= int(loadPct)
	} else {
		score -= node.ActiveTasks * 2
	}

	return score
}
