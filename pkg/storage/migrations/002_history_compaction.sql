-- Migration: 002_history_compaction
CREATE TABLE IF NOT EXISTS run_events (
    id VARCHAR(64) PRIMARY KEY,
    run_id VARCHAR(64) NOT NULL,
    step_id VARCHAR(64),
    tenant_id VARCHAR(64) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    payload_json JSONB
);

CREATE INDEX IF NOT EXISTS idx_events_run ON run_events (run_id, timestamp);

CREATE TABLE IF NOT EXISTS run_history_archives (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    run_id VARCHAR(64) NOT NULL,
    archived_at TIMESTAMP WITH TIME ZONE NOT NULL,
    run_summary JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_archives_tenant ON run_history_archives (tenant_id, archived_at);
