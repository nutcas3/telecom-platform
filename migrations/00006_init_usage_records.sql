-- +goose Up
-- Create usage_records table
CREATE TABLE IF NOT EXISTS usage_records (
    id           SERIAL PRIMARY KEY,
    subscriber_id INTEGER NOT NULL,
    session_id   INTEGER,
    usage_type   TEXT NOT NULL,
    start_time   TIMESTAMPTZ NOT NULL,
    end_time     TIMESTAMPTZ NOT NULL,
    volume       BIGINT NOT NULL,
    rate         DOUBLE PRECISION,
    cost         DOUBLE PRECISION,
    billing_cycle TEXT,
    invoice_id   INTEGER,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_usage_records_subscriber_id ON usage_records(subscriber_id);
CREATE INDEX IF NOT EXISTS idx_usage_records_session_id ON usage_records(session_id);
CREATE INDEX IF NOT EXISTS idx_usage_records_usage_type ON usage_records(usage_type);
CREATE INDEX IF NOT EXISTS idx_usage_records_billing_cycle ON usage_records(billing_cycle);

-- +goose Down
-- Drop usage_records table
DROP TABLE IF EXISTS usage_records CASCADE;
