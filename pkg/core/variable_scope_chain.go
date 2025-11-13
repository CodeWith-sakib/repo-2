package core

import (
	"fmt"
	"sync"
)

// VariableScope represents a single frame in the lexical variable hierarchy.
type VariableScope struct {
	mu        sync.RWMutex
	id        string
	parent    *VariableScope
	bindings  map[string]interface{}
	immutable bool
}

// NewRootScope initializes a global root scope frame with no parent.
func NewRootScope(id string) *VariableScope {
	return &VariableScope{
		id:       id,
		bindings: make(map[string]interface{}),
	}
}

// NewChildScope creates a child scope that inherits and can shadow variables from this parent.
func (s *VariableScope) NewChildScope(childID string) *VariableScope {
	return &VariableScope{
		id:       childID,
		parent:   s,
		bindings: make(map[string]interface{}),
	}
}

// Set binds or overwrites a variable in this local scope frame.
func (s *VariableScope) Set(key string, val interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.immutable {
		return fmt.Errorf("cannot write to immutable scope %s", s.id)
	}
	s.bindings[key] = val
	return nil
}

// Get resolves a variable by searching outward from this scope up through ancestors.
func (s *VariableScope) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	val, ok := s.bindings[key]
	s.mu.RUnlock()

	if ok {
		return val, true
	}

	if s.parent != nil {
		return s.parent.Get(key)
	}

	return nil, false
}

// Flatten extracts all visible variables in this scope into a single merged map.
// Local variables take precedence over ancestor variables (shadowing).
func (s *VariableScope) Flatten() map[string]interface{} {
	out := make(map[string]interface{})

	// Gather ancestors from root down to this scope
	var chain []*VariableScope
	curr := s
	for curr != nil {
		chain = append([]*VariableScope{curr}, chain...)
		curr = curr.parent
	}

	for _, frame := range chain {
		frame.mu.RLock()
		for k, v := range frame.bindings {
			out[k] = v
		}
		frame.mu.RUnlock()
	}

	return out
}

// Depth returns the nesting depth of this scope (root = 0).
func (s *VariableScope) Depth() int {
	depth := 0
	curr := s.parent
	for curr != nil {
		depth++
		curr = curr.parent
	}
	return depth
}

// Freeze marks this scope frame as read-only.
func (s *VariableScope) Freeze() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.immutable = true
}
