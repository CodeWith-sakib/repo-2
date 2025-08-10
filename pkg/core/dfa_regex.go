package core

import (
	"fmt"
	"sync"
)

// StateType characterizes the DFA state
type StateType int

const (
	StateStandard StateType = iota
	StateAccepting
	StateRejecting
)

// DFAState represents an internal state in an expression DFA.
type DFAState struct {
	ID          int
	Type        StateType
	Transitions map[rune]int
}

// DFAMachine executes deterministic finite automata matches on workflow variable expressions.
type DFAMachine struct {
	mu           sync.RWMutex
	states       map[int]*DFAState
	startStateID int
}

// NewDFAMachine initializes an empty DFA engine.
func NewDFAMachine() *DFAMachine {
	return &DFAMachine{
		states: make(map[int]*DFAState),
	}
}

// AddState adds a state to the DFA.
func (d *DFAMachine) AddState(id int, stateType StateType) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.states[id]; !exists {
		d.states[id] = &DFAState{
			ID:          id,
			Type:        stateType,
			Transitions: make(map[rune]int),
		}
	}
}

// SetStartState defines the start state.
func (d *DFAMachine) SetStartState(id int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.states[id]; !exists {
		return fmt.Errorf("state %d does not exist", id)
	}
	d.startStateID = id
	return nil
}

// AddTransition adds a deterministic transition on a rune.
func (d *DFAMachine) AddTransition(fromID int, symbol rune, toID int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	from, ok1 := d.states[fromID]
	if !ok1 {
		return fmt.Errorf("origin state %d missing", fromID)
	}
	if _, ok2 := d.states[toID]; !ok2 {
		return fmt.Errorf("target state %d missing", toID)
	}
	from.Transitions[symbol] = toID
	return nil
}

// Match tests whether a given input string drives the DFA to an accepting state.
func (d *DFAMachine) Match(input string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	currentID := d.startStateID
	for _, r := range input {
		curr, exists := d.states[currentID]
		if !exists {
			return false
		}
		nextID, ok := curr.Transitions[r]
		if !ok {
			return false
		}
		currentID = nextID
	}

	state, ok := d.states[currentID]
	if !ok {
		return false
	}
	return state.Type == StateAccepting
}

// CompileIdentifierPattern creates a DFA matching standard identifier tokens [a-zA-Z_][a-zA-Z0-9_]*
func CompileIdentifierPattern() (*DFAMachine, error) {
	m := NewDFAMachine()
	m.AddState(0, StateStandard)  // start
	m.AddState(1, StateAccepting) // identifier body
	if err := m.SetStartState(0); err != nil {
		return nil, err
	}

	// 0 -> 1 on letters and underscore
	for r := 'a'; r <= 'z'; r++ {
		_ = m.AddTransition(0, r, 1)
		_ = m.AddTransition(1, r, 1)
	}
	for r := 'A'; r <= 'Z'; r++ {
		_ = m.AddTransition(0, r, 1)
		_ = m.AddTransition(1, r, 1)
	}
	_ = m.AddTransition(0, '_', 1)
	_ = m.AddTransition(1, '_', 1)

	for r := '0'; r <= '9'; r++ {
		_ = m.AddTransition(1, r, 1)
	}

	return m, nil
}
