-- +goose Up
-- Create automations table
CREATE TABLE IF NOT EXISTS automations (
    id                  SERIAL PRIMARY KEY,
    name                TEXT NOT NULL UNIQUE,
    description         TEXT,
    type                TEXT,
    enabled             BOOLEAN NOT NULL DEFAULT true,
    schedule_type       TEXT,
    schedule_cron       TEXT,
    schedule_interval_sec INTEGER,
    timezone            TEXT NOT NULL DEFAULT 'UTC',
    definition          JSONB,
    last_run_at         TIMESTAMPTZ,
    next_run_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

-- Create automation_runs table
CREATE TABLE IF NOT EXISTS automation_runs (
    id           SERIAL PRIMARY KEY,
    automation_id INTEGER NOT NULL REFERENCES automations(id) ON DELETE CASCADE,
    status       TEXT NOT NULL,
    started_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at     TIMESTAMPTZ,
    duration_ms  BIGINT NOT NULL DEFAULT 0,
    output       TEXT,
    error        TEXT,
    details      JSONB,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_automations_name ON automations(name);
CREATE INDEX IF NOT EXISTS idx_automations_enabled ON automations(enabled);
CREATE INDEX IF NOT EXISTS idx_automations_deleted_at ON automations(deleted_at);
CREATE INDEX IF NOT EXISTS idx_automation_runs_automation_id ON automation_runs(automation_id);
CREATE INDEX IF NOT EXISTS idx_automation_runs_status ON automation_runs(status);
CREATE INDEX IF NOT EXISTS idx_automation_runs_started_at ON automation_runs(started_at);

-- +goose Down
-- Drop automation_runs and automations tables
DROP TABLE IF EXISTS automation_runs CASCADE;
DROP TABLE IF EXISTS automations CASCADE;
