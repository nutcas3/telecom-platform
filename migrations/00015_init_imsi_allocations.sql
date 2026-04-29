-- +goose Up
-- Create imsi_allocations table
CREATE TABLE IF NOT EXISTS imsi_allocations (
    id        SERIAL PRIMARY KEY,
    last_imsi BIGINT NOT NULL DEFAULT 0,
    min_imsi  BIGINT NOT NULL DEFAULT 1,
    max_imsi  BIGINT NOT NULL DEFAULT 999999999,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
-- Drop imsi_allocations table
DROP TABLE IF EXISTS imsi_allocations CASCADE;
