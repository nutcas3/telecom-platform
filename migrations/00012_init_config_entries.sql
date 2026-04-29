-- +goose Up
-- Create config_entries table
CREATE TABLE IF NOT EXISTS config_entries (
    id          SERIAL PRIMARY KEY,
    section     TEXT NOT NULL,
    key         TEXT NOT NULL,
    value       TEXT,
    type        TEXT,
    sensitive   BOOLEAN NOT NULL DEFAULT false,
    description TEXT,
    updated_by  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(section, key)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_config_entries_section ON config_entries(section);
CREATE INDEX IF NOT EXISTS idx_config_entries_key ON config_entries(key);

-- +goose Down
-- Drop config_entries table
DROP TABLE IF EXISTS config_entries CASCADE;
