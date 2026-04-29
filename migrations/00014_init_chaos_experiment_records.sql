-- +goose Up
-- Create chaos_experiment_records table
CREATE TABLE IF NOT EXISTS chaos_experiment_records (
    id          SERIAL PRIMARY KEY,
    external_id TEXT UNIQUE,
    name        TEXT,
    type        TEXT,
    target      TEXT,
    status      TEXT,
    duration_ms BIGINT NOT NULL DEFAULT 0,
    probability DOUBLE PRECISION,
    amount      INTEGER NOT NULL DEFAULT 0,
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ,
    error       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_chaos_experiment_records_external_id ON chaos_experiment_records(external_id);
CREATE INDEX IF NOT EXISTS idx_chaos_experiment_records_status ON chaos_experiment_records(status);
CREATE INDEX IF NOT EXISTS idx_chaos_experiment_records_target ON chaos_experiment_records(target);
CREATE INDEX IF NOT EXISTS idx_chaos_experiment_records_started_at ON chaos_experiment_records(started_at);

-- +goose Down
-- Drop chaos_experiment_records table
DROP TABLE IF EXISTS chaos_experiment_records CASCADE;
