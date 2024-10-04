package postgres

const (
	insertWorkflowSQL = `
		INSERT INTO workflows (id, tenant_id, name, version, description, schema_json, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	getWorkflowSQL = `
		SELECT id, tenant_id, name, version, description, schema_json, created_at, updated_at
		FROM workflows
		WHERE id = $1
		ORDER BY version DESC
		LIMIT 1
	`
	getWorkflowVersionSQL = `
		SELECT id, tenant_id, name, version, description, schema_json, created_at, updated_at
		FROM workflows
		WHERE id = $1 AND version = $2
	`
	insertRunSQL = `
		INSERT INTO workflow_runs (id, workflow_id, version, tenant_id, state, input_json, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	getRunSQL = `
		SELECT id, workflow_id, version, tenant_id, state, input_json, output_json, error_message, priority, started_at, finished_at, created_at, updated_at
		FROM workflow_runs
		WHERE id = $1
	`
	updateRunSQL = `
		UPDATE workflow_runs
		SET state = $2, output_json = $3, error_message = $4, started_at = $5, finished_at = $6, updated_at = $7
		WHERE id = $1
	`
	insertStepRunSQL = `
		INSERT INTO step_runs (id, run_id, step_id, state, attempt, input_json, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	enqueueTaskSQL = `
		INSERT INTO task_queue (id, run_id, step_id, tenant_id, priority, attempt, scheduled_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	ackTaskSQL = `
		DELETE FROM task_queue
		WHERE id = $1 AND lease_worker = $2
	`
)
