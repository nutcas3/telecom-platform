-- +goose Up
-- Create plugins table
CREATE TABLE IF NOT EXISTS plugins (
    id          SERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    version     TEXT NOT NULL,
    description TEXT,
    author      TEXT,
    type        TEXT,
    category    TEXT,
    license     TEXT,
    homepage    TEXT,
    repository  TEXT,
    enabled     BOOLEAN NOT NULL DEFAULT false,
    status      TEXT NOT NULL DEFAULT 'installed',
    config      JSONB,
    installed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_plugins_name ON plugins(name);
CREATE INDEX IF NOT EXISTS idx_plugins_enabled ON plugins(enabled);
CREATE INDEX IF NOT EXISTS idx_plugins_deleted_at ON plugins(deleted_at);

-- +goose Down
-- Drop plugins table
DROP TABLE IF EXISTS plugins CASCADE;
