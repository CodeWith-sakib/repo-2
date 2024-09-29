-- Migration: 001_initial_schema
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workflows (
    id VARCHAR(64) NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    version INT NOT NULL,
    description TEXT,
    schema_json JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (id, version)
);

CREATE INDEX IF NOT EXISTS idx_workflows_tenant ON workflows (tenant_id);

CREATE TABLE IF NOT EXISTS workflow_runs (
    id VARCHAR(64) PRIMARY KEY,
    workflow_id VARCHAR(64) NOT NULL,
    version INT NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    state VARCHAR(32) NOT NULL,
    input_json JSONB,
    output_json JSONB,
    error_message TEXT,
    priority INT DEFAULT 50,
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_runs_tenant_state ON workflow_runs (tenant_id, state);
CREATE INDEX IF NOT EXISTS idx_runs_workflow ON workflow_runs (workflow_id);

CREATE TABLE IF NOT EXISTS step_runs (
    id VARCHAR(64) PRIMARY KEY,
    run_id VARCHAR(64) NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    step_id VARCHAR(64) NOT NULL,
    state VARCHAR(32) NOT NULL,
    attempt INT DEFAULT 1,
    worker_id VARCHAR(64),
    lease_until TIMESTAMP WITH TIME ZONE,
    input_json JSONB,
    output_json JSONB,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_step_runs_run_step ON step_runs (run_id, step_id);

CREATE TABLE IF NOT EXISTS task_queue (
    id VARCHAR(64) PRIMARY KEY,
    run_id VARCHAR(64) NOT NULL,
    step_id VARCHAR(64) NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    priority INT DEFAULT 50,
    attempt INT DEFAULT 1,
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    lease_worker VARCHAR(64),
    lease_until TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_queue_poll ON task_queue (scheduled_at, priority DESC) WHERE lease_worker IS NULL;
