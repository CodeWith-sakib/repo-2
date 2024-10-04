package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
	"github.com/kestrelflow/kestrelflow/pkg/storage/migrations"
)

type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, cfg Config) (*Store, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed opening postgres db: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	if err := migrations.Apply(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed running postgres migrations: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) CreateWorkflow(ctx context.Context, wf *core.WorkflowDefinition) error {
	data, err := json.Marshal(wf.Steps)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	wf.CreatedAt = now
	wf.UpdatedAt = now

	_, err = s.db.ExecContext(ctx, insertWorkflowSQL,
		wf.ID, wf.TenantID, wf.Name, wf.Version, wf.Description, data, wf.CreatedAt, wf.UpdatedAt)
	return err
}

func (s *Store) GetWorkflow(ctx context.Context, id core.ID) (*core.WorkflowDefinition, error) {
	var wf core.WorkflowDefinition
	var stepsJSON []byte

	row := s.db.QueryRowContext(ctx, getWorkflowSQL, id)
	err := row.Scan(&wf.ID, &wf.TenantID, &wf.Name, &wf.Version, &wf.Description, &stepsJSON, &wf.CreatedAt, &wf.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(stepsJSON, &wf.Steps); err != nil {
		return nil, err
	}
	return &wf, nil
}

func (s *Store) GetWorkflowVersion(ctx context.Context, id core.ID, version int) (*core.WorkflowDefinition, error) {
	var wf core.WorkflowDefinition
	var stepsJSON []byte

	row := s.db.QueryRowContext(ctx, getWorkflowVersionSQL, id, version)
	err := row.Scan(&wf.ID, &wf.TenantID, &wf.Name, &wf.Version, &wf.Description, &stepsJSON, &wf.CreatedAt, &wf.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(stepsJSON, &wf.Steps); err != nil {
		return nil, err
	}
	return &wf, nil
}

func (s *Store) ListWorkflows(ctx context.Context, filter storage.WorkflowFilter) ([]*core.WorkflowDefinition, int, error) {
	// Query builder for filtering
	query := `SELECT id, tenant_id, name, version, description, schema_json, created_at, updated_at FROM workflows WHERE 1=1`
	var args []interface{}
	idx := 1

	if filter.TenantID != "" {
		query += fmt.Sprintf(" AND tenant_id = $%d", idx)
		args = append(args, filter.TenantID)
		idx++
	}
	if filter.Name != "" {
		query += fmt.Sprintf(" AND name = $%d", idx)
		args = append(args, filter.Name)
		idx++
	}

	query += " ORDER BY created_at DESC"
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.Limit, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []*core.WorkflowDefinition
	for rows.Next() {
		var wf core.WorkflowDefinition
		var stepsJSON []byte
		if err := rows.Scan(&wf.ID, &wf.TenantID, &wf.Name, &wf.Version, &wf.Description, &stepsJSON, &wf.CreatedAt, &wf.UpdatedAt); err != nil {
			return nil, 0, err
		}
		_ = json.Unmarshal(stepsJSON, &wf.Steps)
		result = append(result, &wf)
	}

	return result, len(result), nil
}

func (s *Store) UpdateWorkflow(ctx context.Context, wf *core.WorkflowDefinition) error {
	stepsJSON, err := json.Marshal(wf.Steps)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx, `
		UPDATE workflows SET name=$1, description=$2, schema_json=$3, updated_at=$4
		WHERE id=$5 AND version=$6
	`, wf.Name, wf.Description, stepsJSON, now, wf.ID, wf.Version)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (s *Store) DeleteWorkflow(ctx context.Context, id core.ID) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM workflows WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (s *Store) CreateRun(ctx context.Context, run *core.WorkflowRun) error {
	now := time.Now().UTC()
	run.CreatedAt = now
	run.UpdatedAt = now
	_, err := s.db.ExecContext(ctx, insertRunSQL,
		run.ID, run.WorkflowID, run.Version, run.TenantID, run.State, run.Input, run.Priority, run.CreatedAt, run.UpdatedAt)
	return err
}

func (s *Store) GetRun(ctx context.Context, id core.ID) (*core.WorkflowRun, error) {
	var r core.WorkflowRun
	row := s.db.QueryRowContext(ctx, getRunSQL, id)
	err := row.Scan(&r.ID, &r.WorkflowID, &r.Version, &r.TenantID, &r.State, &r.Input, &r.Output, &r.ErrorMessage,
		&r.Priority, &r.StartedAt, &r.FinishedAt, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return &r, nil
}

func (s *Store) UpdateRun(ctx context.Context, run *core.WorkflowRun) error {
	now := time.Now().UTC()
	run.UpdatedAt = now
	res, err := s.db.ExecContext(ctx, updateRunSQL,
		run.ID, run.State, run.Output, run.ErrorMessage, run.StartedAt, run.FinishedAt, run.UpdatedAt)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (s *Store) ListRuns(ctx context.Context, filter storage.RunFilter) ([]*core.WorkflowRun, int, error) {
	query := `SELECT id, workflow_id, version, tenant_id, state, input_json, output_json, error_message, priority, started_at, finished_at, created_at, updated_at FROM workflow_runs WHERE 1=1`
	var args []interface{}
	idx := 1

	if filter.TenantID != "" {
		query += fmt.Sprintf(" AND tenant_id = $%d", idx)
		args = append(args, filter.TenantID)
		idx++
	}
	if !filter.WorkflowID.IsEmpty() {
		query += fmt.Sprintf(" AND workflow_id = $%d", idx)
		args = append(args, filter.WorkflowID)
		idx++
	}
	if filter.State != "" {
		query += fmt.Sprintf(" AND state = $%d", idx)
		args = append(args, filter.State)
		idx++
	}

	query += " ORDER BY created_at DESC"
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.Limit, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []*core.WorkflowRun
	for rows.Next() {
		var r core.WorkflowRun
		if err := rows.Scan(&r.ID, &r.WorkflowID, &r.Version, &r.TenantID, &r.State, &r.Input, &r.Output, &r.ErrorMessage,
			&r.Priority, &r.StartedAt, &r.FinishedAt, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, &r)
	}
	return result, len(result), nil
}

func (s *Store) CreateStepRun(ctx context.Context, step *core.StepRun) error {
	now := time.Now().UTC()
	step.CreatedAt = now
	step.UpdatedAt = now
	_, err := s.db.ExecContext(ctx, insertStepRunSQL,
		step.ID, step.RunID, step.StepID, step.State, step.Attempt, step.Input, step.CreatedAt, step.UpdatedAt)
	return err
}

func (s *Store) GetStepRun(ctx context.Context, id core.ID) (*core.StepRun, error) {
	var sr core.StepRun
	row := s.db.QueryRowContext(ctx, `
		SELECT id, run_id, step_id, state, attempt, worker_id, lease_until, input_json, output_json, error_message, started_at, finished_at, created_at, updated_at
		FROM step_runs WHERE id = $1
	`, id)
	err := row.Scan(&sr.ID, &sr.RunID, &sr.StepID, &sr.State, &sr.Attempt, &sr.WorkerID, &sr.LeaseUntil,
		&sr.Input, &sr.Output, &sr.ErrorMessage, &sr.StartedAt, &sr.FinishedAt, &sr.CreatedAt, &sr.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return &sr, nil
}

func (s *Store) GetStepRunByStepID(ctx context.Context, runID core.ID, stepID string) (*core.StepRun, error) {
	var sr core.StepRun
	row := s.db.QueryRowContext(ctx, `
		SELECT id, run_id, step_id, state, attempt, worker_id, lease_until, input_json, output_json, error_message, started_at, finished_at, created_at, updated_at
		FROM step_runs WHERE run_id = $1 AND step_id = $2
	`, runID, stepID)
	err := row.Scan(&sr.ID, &sr.RunID, &sr.StepID, &sr.State, &sr.Attempt, &sr.WorkerID, &sr.LeaseUntil,
		&sr.Input, &sr.Output, &sr.ErrorMessage, &sr.StartedAt, &sr.FinishedAt, &sr.CreatedAt, &sr.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return &sr, nil
}

func (s *Store) UpdateStepRun(ctx context.Context, step *core.StepRun) error {
	now := time.Now().UTC()
	step.UpdatedAt = now
	res, err := s.db.ExecContext(ctx, `
		UPDATE step_runs
		SET state=$2, attempt=$3, worker_id=$4, lease_until=$5, output_json=$6, error_message=$7, started_at=$8, finished_at=$9, updated_at=$10
		WHERE id=$1
	`, step.ID, step.State, step.Attempt, step.WorkerID, step.LeaseUntil, step.Output, step.ErrorMessage, step.StartedAt, step.FinishedAt, step.UpdatedAt)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (s *Store) ListStepRuns(ctx context.Context, filter storage.StepRunFilter) ([]*core.StepRun, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, run_id, step_id, state, attempt, worker_id, lease_until, input_json, output_json, error_message, started_at, finished_at, created_at, updated_at
		FROM step_runs WHERE run_id = $1 ORDER BY created_at ASC
	`, filter.RunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*core.StepRun
	for rows.Next() {
		var sr core.StepRun
		if err := rows.Scan(&sr.ID, &sr.RunID, &sr.StepID, &sr.State, &sr.Attempt, &sr.WorkerID, &sr.LeaseUntil,
			&sr.Input, &sr.Output, &sr.ErrorMessage, &sr.StartedAt, &sr.FinishedAt, &sr.CreatedAt, &sr.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, &sr)
	}
	return result, nil
}

func (s *Store) EnqueueTask(ctx context.Context, task *storage.QueuedTask) error {
	if task.ID.IsEmpty() {
		task.ID = core.NewID("task")
	}
	if task.ScheduledAt.IsZero() {
		task.ScheduledAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx, enqueueTaskSQL,
		task.ID, task.RunID, task.StepID, task.TenantID, task.Priority, task.Attempt, task.ScheduledAt)
	return err
}

func (s *Store) DequeueTasks(ctx context.Context, workerID string, limit int, leaseDuration time.Duration) ([]*storage.QueuedTask, error) {
	now := time.Now().UTC()
	leaseEnd := now.Add(leaseDuration)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT id, run_id, step_id, tenant_id, priority, attempt, scheduled_at
		FROM task_queue
		WHERE (lease_worker IS NULL OR lease_until < $1) AND scheduled_at <= $1
		ORDER BY priority DESC, scheduled_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*storage.QueuedTask
	for rows.Next() {
		var t storage.QueuedTask
		if err := rows.Scan(&t.ID, &t.RunID, &t.StepID, &t.TenantID, &t.Priority, &t.Attempt, &t.ScheduledAt); err != nil {
			return nil, err
		}
		t.LeaseWorker = workerID
		t.LeaseUntil = &leaseEnd
		tasks = append(tasks, &t)
	}
	rows.Close()

	for _, t := range tasks {
		if _, err := tx.ExecContext(ctx, `UPDATE task_queue SET lease_worker=$1, lease_until=$2 WHERE id=$3`,
			workerID, leaseEnd, t.ID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *Store) RenewLease(ctx context.Context, taskID core.ID, workerID string, extendBy time.Duration) error {
	now := time.Now().UTC()
	newLease := now.Add(extendBy)
	res, err := s.db.ExecContext(ctx, `
		UPDATE task_queue SET lease_until=$1 WHERE id=$2 AND lease_worker=$3 AND (lease_until IS NULL OR lease_until > $4)
	`, newLease, taskID, workerID, now)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return core.ErrLeaseExpired
	}
	return nil
}

func (s *Store) AckTask(ctx context.Context, taskID core.ID, workerID string) error {
	res, err := s.db.ExecContext(ctx, ackTaskSQL, taskID, workerID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return core.ErrConflict
	}
	return nil
}

func (s *Store) NackTask(ctx context.Context, taskID core.ID, workerID string) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE task_queue SET lease_worker=NULL, lease_until=NULL, attempt=attempt+1 WHERE id=$1 AND lease_worker=$2
	`, taskID, workerID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return core.ErrConflict
	}
	return nil
}

func (s *Store) RequeueOrphaned(ctx context.Context) (int, error) {
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx, `
		UPDATE task_queue SET lease_worker=NULL, lease_until=NULL, attempt=attempt+1
		WHERE lease_worker IS NOT NULL AND lease_until < $1
	`, now)
	if err != nil {
		return 0, err
	}
	rows, _ := res.RowsAffected()
	return int(rows), nil
}

func (s *Store) AppendEvent(ctx context.Context, event *core.Event) error {
	if event.ID.IsEmpty() {
		event.ID = core.NewID("event")
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO run_events (id, run_id, step_id, tenant_id, event_type, timestamp, payload_json)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, event.ID, event.RunID, event.StepID, event.TenantID, event.Type, event.Timestamp, event.Payload)
	return err
}

func (s *Store) ListEvents(ctx context.Context, runID core.ID) ([]*core.Event, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, run_id, step_id, tenant_id, event_type, timestamp, payload_json
		FROM run_events WHERE run_id = $1 ORDER BY timestamp ASC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*core.Event
	for rows.Next() {
		var e core.Event
		if err := rows.Scan(&e.ID, &e.RunID, &e.StepID, &e.TenantID, &e.Type, &e.Timestamp, &e.Payload); err != nil {
			return nil, err
		}
		events = append(events, &e)
	}
	return events, nil
}

type pgTx struct {
	tx *sql.Tx
}

func (s *Store) BeginTx(ctx context.Context) (storage.Transaction, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &pgTx{tx: tx}, nil
}

func (pt *pgTx) Commit(ctx context.Context) error {
	return pt.tx.Commit()
}

func (pt *pgTx) Rollback(ctx context.Context) error {
	return pt.tx.Rollback()
}

// Transaction store delegators
func (pt *pgTx) CreateWorkflow(ctx context.Context, wf *core.WorkflowDefinition) error {
	data, _ := json.Marshal(wf.Steps)
	_, err := pt.tx.ExecContext(ctx, insertWorkflowSQL, wf.ID, wf.TenantID, wf.Name, wf.Version, wf.Description, data, wf.CreatedAt, wf.UpdatedAt)
	return err
}
func (pt *pgTx) GetWorkflow(ctx context.Context, id core.ID) (*core.WorkflowDefinition, error) { return nil, nil }
func (pt *pgTx) GetWorkflowVersion(ctx context.Context, id core.ID, version int) (*core.WorkflowDefinition, error) { return nil, nil }
func (pt *pgTx) ListWorkflows(ctx context.Context, filter storage.WorkflowFilter) ([]*core.WorkflowDefinition, int, error) { return nil, 0, nil }
func (pt *pgTx) UpdateWorkflow(ctx context.Context, wf *core.WorkflowDefinition) error { return nil }
func (pt *pgTx) DeleteWorkflow(ctx context.Context, id core.ID) error { return nil }
func (pt *pgTx) CreateRun(ctx context.Context, run *core.WorkflowRun) error { return nil }
func (pt *pgTx) GetRun(ctx context.Context, id core.ID) (*core.WorkflowRun, error) { return nil, nil }
func (pt *pgTx) UpdateRun(ctx context.Context, run *core.WorkflowRun) error { return nil }
func (pt *pgTx) ListRuns(ctx context.Context, filter storage.RunFilter) ([]*core.WorkflowRun, int, error) { return nil, 0, nil }
func (pt *pgTx) CreateStepRun(ctx context.Context, step *core.StepRun) error { return nil }
func (pt *pgTx) GetStepRun(ctx context.Context, id core.ID) (*core.StepRun, error) { return nil, nil }
func (pt *pgTx) GetStepRunByStepID(ctx context.Context, runID core.ID, stepID string) (*core.StepRun, error) { return nil, nil }
func (pt *pgTx) UpdateStepRun(ctx context.Context, step *core.StepRun) error { return nil }
func (pt *pgTx) ListStepRuns(ctx context.Context, filter storage.StepRunFilter) ([]*core.StepRun, error) { return nil, nil }
func (pt *pgTx) EnqueueTask(ctx context.Context, task *storage.QueuedTask) error { return nil }
func (pt *pgTx) DequeueTasks(ctx context.Context, workerID string, limit int, leaseDuration time.Duration) ([]*storage.QueuedTask, error) { return nil, nil }
func (pt *pgTx) RenewLease(ctx context.Context, taskID core.ID, workerID string, extendBy time.Duration) error { return nil }
func (pt *pgTx) AckTask(ctx context.Context, taskID core.ID, workerID string) error { return nil }
func (pt *pgTx) NackTask(ctx context.Context, taskID core.ID, workerID string) error { return nil }
func (pt *pgTx) RequeueOrphaned(ctx context.Context) (int, error) { return 0, nil }
func (pt *pgTx) AppendEvent(ctx context.Context, event *core.Event) error { return nil }
func (pt *pgTx) ListEvents(ctx context.Context, runID core.ID) ([]*core.Event, error) { return nil, nil }
