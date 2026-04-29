-- +goose Up
-- Create service_plans table
CREATE TABLE IF NOT EXISTS service_plans (
    id          SERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT,
    data_limit  BIGINT,
    voice_limit INTEGER,
    sms_limit   INTEGER,
    monthly_fee DOUBLE PRECISION,
    data_rate   DOUBLE PRECISION,
    voice_rate  DOUBLE PRECISION,
    sms_rate    DOUBLE PRECISION,
    arpa        INTEGER,
    ambr_uplink   INTEGER,
    ambr_downlink INTEGER,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_service_plans_name ON service_plans(name);
CREATE INDEX IF NOT EXISTS idx_service_plans_deleted_at ON service_plans(deleted_at);

-- +goose Down
-- Drop service_plans table
DROP TABLE IF EXISTS service_plans CASCADE;
