package memory

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

type Store struct {
	mu           sync.RWMutex
	workflows    map[core.ID]map[int]*core.WorkflowDefinition // ID -> Version -> Def
	runs         map[core.ID]*core.WorkflowRun
	stepRuns     map[core.ID]*core.StepRun
	runSteps     map[core.ID][]core.ID // runID -> stepRunIDs
	queuedTasks  map[core.ID]*storage.QueuedTask
	events       map[core.ID][]*core.Event // runID -> events
}

func NewStore() *Store {
	return &Store{
		workflows:   make(map[core.ID]map[int]*core.WorkflowDefinition),
		runs:        make(map[core.ID]*core.WorkflowRun),
		stepRuns:    make(map[core.ID]*core.StepRun),
		runSteps:    make(map[core.ID][]core.ID),
		queuedTasks: make(map[core.ID]*storage.QueuedTask),
		events:      make(map[core.ID][]*core.Event),
	}
}

func (s *Store) Close() error {
	return nil
}

func cloneWorkflow(w *core.WorkflowDefinition) *core.WorkflowDefinition {
	if w == nil {
		return nil
	}
	data, _ := json.Marshal(w)
	var clone core.WorkflowDefinition
	_ = json.Unmarshal(data, &clone)
	return &clone
}

func cloneRun(r *core.WorkflowRun) *core.WorkflowRun {
	if r == nil {
		return nil
	}
	data, _ := json.Marshal(r)
	var clone core.WorkflowRun
	_ = json.Unmarshal(data, &clone)
	return &clone
}

func cloneStepRun(sr *core.StepRun) *core.StepRun {
	if sr == nil {
		return nil
	}
	data, _ := json.Marshal(sr)
	var clone core.StepRun
	_ = json.Unmarshal(data, &clone)
	return &clone
}

func (s *Store) CreateWorkflow(ctx context.Context, wf *core.WorkflowDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.workflows[wf.ID]; !exists {
		s.workflows[wf.ID] = make(map[int]*core.WorkflowDefinition)
	}

	if _, exists := s.workflows[wf.ID][wf.Version]; exists {
		return core.ErrAlreadyExists
	}

	wf.CreatedAt = time.Now().UTC()
	wf.UpdatedAt = wf.CreatedAt
	s.workflows[wf.ID][wf.Version] = cloneWorkflow(wf)
	return nil
}

func (s *Store) GetWorkflow(ctx context.Context, id core.ID) (*core.WorkflowDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	versions, exists := s.workflows[id]
	if !exists || len(versions) == 0 {
		return nil, core.ErrNotFound
	}

	maxVer := 0
	var latest *core.WorkflowDefinition
	for ver, def := range versions {
		if ver > maxVer {
			maxVer = ver
			latest = def
		}
	}
	return cloneWorkflow(latest), nil
}

func (s *Store) GetWorkflowVersion(ctx context.Context, id core.ID, version int) (*core.WorkflowDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	versions, exists := s.workflows[id]
	if !exists {
		return nil, core.ErrNotFound
	}

	def, exists := versions[version]
	if !exists {
		return nil, core.ErrNotFound
	}
	return cloneWorkflow(def), nil
}

func (s *Store) ListWorkflows(ctx context.Context, filter storage.WorkflowFilter) ([]*core.WorkflowDefinition, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*core.WorkflowDefinition
	for _, versions := range s.workflows {
		maxVer := 0
		var latest *core.WorkflowDefinition
		for ver, def := range versions {
			if ver > maxVer {
				maxVer = ver
				latest = def
			}
		}
		if latest == nil {
			continue
		}
		if filter.TenantID != "" && latest.TenantID != filter.TenantID {
			continue
		}
		if filter.Name != "" && latest.Name != filter.Name {
			continue
		}
		result = append(result, cloneWorkflow(latest))
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	total := len(result)
	start := filter.Offset
	if start > total {
		return []*core.WorkflowDefinition{}, total, nil
	}
	end := start + filter.Limit
	if filter.Limit <= 0 || end > total {
		end = total
	}
	return result[start:end], total, nil
}

func (s *Store) UpdateWorkflow(ctx context.Context, wf *core.WorkflowDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	versions, exists := s.workflows[wf.ID]
	if !exists {
		return core.ErrNotFound
	}
	if _, exists := versions[wf.Version]; !exists {
		return core.ErrNotFound
	}

	wf.UpdatedAt = time.Now().UTC()
	versions[wf.Version] = cloneWorkflow(wf)
	return nil
}

func (s *Store) DeleteWorkflow(ctx context.Context, id core.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.workflows[id]; !exists {
		return core.ErrNotFound
	}
	delete(s.workflows, id)
	return nil
}

func (s *Store) CreateRun(ctx context.Context, run *core.WorkflowRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.runs[run.ID]; exists {
		return core.ErrAlreadyExists
	}
	now := time.Now().UTC()
	run.CreatedAt = now
	run.UpdatedAt = now
	s.runs[run.ID] = cloneRun(run)
	return nil
}

func (s *Store) GetRun(ctx context.Context, id core.ID) (*core.WorkflowRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	run, exists := s.runs[id]
	if !exists {
		return nil, core.ErrNotFound
	}
	return cloneRun(run), nil
}

func (s *Store) UpdateRun(ctx context.Context, run *core.WorkflowRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.runs[run.ID]; !exists {
		return core.ErrNotFound
	}
	run.UpdatedAt = time.Now().UTC()
	s.runs[run.ID] = cloneRun(run)
	return nil
}

func (s *Store) ListRuns(ctx context.Context, filter storage.RunFilter) ([]*core.WorkflowRun, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matched []*core.WorkflowRun
	for _, r := range s.runs {
		if filter.TenantID != "" && r.TenantID != filter.TenantID {
			continue
		}
		if !filter.WorkflowID.IsEmpty() && r.WorkflowID != filter.WorkflowID {
			continue
		}
		if filter.State != "" && r.State != filter.State {
			continue
		}
		if filter.CreatedAfter != nil && r.CreatedAt.Before(*filter.CreatedAfter) {
			continue
		}
		if filter.CreatedBefore != nil && r.CreatedAt.After(*filter.CreatedBefore) {
			continue
		}
		matched = append(matched, cloneRun(r))
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})

	total := len(matched)
	start := filter.Offset
	if start > total {
		return []*core.WorkflowRun{}, total, nil
	}
	end := start + filter.Limit
	if filter.Limit <= 0 || end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (s *Store) CreateStepRun(ctx context.Context, step *core.StepRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.stepRuns[step.ID]; exists {
		return core.ErrAlreadyExists
	}
	now := time.Now().UTC()
	step.CreatedAt = now
	step.UpdatedAt = now

	s.stepRuns[step.ID] = cloneStepRun(step)
	s.runSteps[step.RunID] = append(s.runSteps[step.RunID], step.ID)
	return nil
}

func (s *Store) GetStepRun(ctx context.Context, id core.ID) (*core.StepRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sr, exists := s.stepRuns[id]
	if !exists {
		return nil, core.ErrNotFound
	}
	return cloneStepRun(sr), nil
}

func (s *Store) GetStepRunByStepID(ctx context.Context, runID core.ID, stepID string) (*core.StepRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stepIDs, exists := s.runSteps[runID]
	if !exists {
		return nil, core.ErrNotFound
	}

	for _, id := range stepIDs {
		sr := s.stepRuns[id]
		if sr != nil && sr.StepID == stepID {
			return cloneStepRun(sr), nil
		}
	}
	return nil, core.ErrNotFound
}

func (s *Store) UpdateStepRun(ctx context.Context, step *core.StepRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.stepRuns[step.ID]; !exists {
		return core.ErrNotFound
	}
	step.UpdatedAt = time.Now().UTC()
	s.stepRuns[step.ID] = cloneStepRun(step)
	return nil
}

func (s *Store) ListStepRuns(ctx context.Context, filter storage.StepRunFilter) ([]*core.StepRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stepIDs, exists := s.runSteps[filter.RunID]
	if !exists {
		return []*core.StepRun{}, nil
	}

	var result []*core.StepRun
	for _, id := range stepIDs {
		sr := s.stepRuns[id]
		if sr == nil {
			continue
		}
		if filter.StepID != "" && sr.StepID != filter.StepID {
			continue
		}
		if filter.State != "" && sr.State != filter.State {
			continue
		}
		result = append(result, cloneStepRun(sr))
	}
	return result, nil
}
