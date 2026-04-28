-- +goose Up
-- Create rating_plans table
CREATE TABLE IF NOT EXISTS rating_plans (
    plan_id      TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    data_rate    DOUBLE PRECISION NOT NULL,
    voice_rate   DOUBLE PRECISION NOT NULL,
    sms_rate     DOUBLE PRECISION NOT NULL,
    monthly_fee  DOUBLE PRECISION NOT NULL,
    data_limit   BIGINT NOT NULL,
    voice_limit  BIGINT NOT NULL,
    sms_limit    BIGINT NOT NULL,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_rating_plans_is_active ON rating_plans(is_active);
CREATE INDEX IF NOT EXISTS idx_rating_plans_created_at ON rating_plans(created_at);

-- +goose Down
-- Drop rating_plans table
DROP TABLE IF EXISTS rating_plans CASCADE;
