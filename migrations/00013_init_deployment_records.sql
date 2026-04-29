-- +goose Up
-- Create deployment_records table
CREATE TABLE IF NOT EXISTS deployment_records (
    id          SERIAL PRIMARY KEY,
    service     TEXT NOT NULL,
    version     TEXT NOT NULL,
    environment TEXT,
    status      TEXT NOT NULL,
    strategy    TEXT,
    replicas    INTEGER NOT NULL DEFAULT 1,
    triggered_by TEXT,
    reason      TEXT,
    started_at  TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    rollback_to TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_deployment_records_service ON deployment_records(service);
CREATE INDEX IF NOT EXISTS idx_deployment_records_environment ON deployment_records(environment);
CREATE INDEX IF NOT EXISTS idx_deployment_records_status ON deployment_records(status);
CREATE INDEX IF NOT EXISTS idx_deployment_records_created_at ON deployment_records(created_at);

-- +goose Down
-- Drop deployment_records table
DROP TABLE IF EXISTS deployment_records CASCADE;
