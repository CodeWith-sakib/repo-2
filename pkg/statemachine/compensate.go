package statemachine

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type CompensationAction struct {
	StepID      string          `json:"step_id"`
	ActionType  string          `json:"action_type"`
	Payload     json.RawMessage `json:"payload"`
	CreatedAt   time.Time       `json:"created_at"`
	ExecutedAt  *time.Time      `json:"executed_at,omitempty"`
	Success     bool            `json:"success"`
	ErrorReason string          `json:"error_reason,omitempty"`
}

type SagaManager struct {
	mu            sync.RWMutex
	compensations map[core.ID][]*CompensationAction
}

func NewSagaManager() *SagaManager {
	return &SagaManager{
		compensations: make(map[core.ID][]*CompensationAction),
	}
}

func (s *SagaManager) RegisterCompensation(runID core.ID, action *CompensationAction) {
	s.mu.Lock()
	defer s.mu.Unlock()

	action.CreatedAt = time.Now().UTC()
	s.compensations[runID] = append(s.compensations[runID], action)
}

func (s *SagaManager) GetCompensations(runID core.ID) []*CompensationAction {
	s.mu.RLock()
	defer s.mu.RUnlock()

	actions := s.compensations[runID]
	res := make([]*CompensationAction, len(actions))
	copy(res, actions)
	return res
}

func (s *SagaManager) Compensate(ctx context.Context, runID core.ID, executor func(ctx context.Context, action *CompensationAction) error) error {
	s.mu.Lock()
	actions, exists := s.compensations[runID]
	s.mu.Unlock()

	if !exists || len(actions) == 0 {
		return nil
	}

	// Reverse order compensation (LIFO)
	for i := len(actions) - 1; i >= 0; i-- {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		action := actions[i]
		now := time.Now().UTC()
		action.ExecutedAt = &now

		err := executor(ctx, action)
		if err != nil {
			action.Success = false
			action.ErrorReason = err.Error()
			return fmt.Errorf("compensation failed for step %s: %w", action.StepID, err)
		}
		action.Success = true
	}

	return nil
}
