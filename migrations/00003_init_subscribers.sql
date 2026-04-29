-- +goose Up
-- Create subscribers table
CREATE TABLE IF NOT EXISTS subscribers (
    id            SERIAL PRIMARY KEY,
    imsi          TEXT NOT NULL UNIQUE,
    msisdn        TEXT UNIQUE,
    imei          TEXT,
    first_name    TEXT NOT NULL,
    last_name     TEXT NOT NULL,
    email         TEXT UNIQUE,
    organization_id TEXT,
    status        TEXT NOT NULL DEFAULT 'active',
    plan_id       INTEGER,
    balance       DOUBLE PRECISION NOT NULL DEFAULT 0,
    auth_key      TEXT NOT NULL,
    opc           TEXT NOT NULL,
    serving_plmn_mcc TEXT NOT NULL,
    serving_plmn_mnc TEXT NOT NULL,
    euicc_id      TEXT,
    profile_id    TEXT,
    profile_status TEXT NOT NULL DEFAULT 'inactive',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_subscribers_imsi ON subscribers(imsi);
CREATE INDEX IF NOT EXISTS idx_subscribers_msisdn ON subscribers(msisdn);
CREATE INDEX IF NOT EXISTS idx_subscribers_email ON subscribers(email);
CREATE INDEX IF NOT EXISTS idx_subscribers_status ON subscribers(status);
CREATE INDEX IF NOT EXISTS idx_subscribers_organization_id ON subscribers(organization_id);
CREATE INDEX IF NOT EXISTS idx_subscribers_plan_id ON subscribers(plan_id);
CREATE INDEX IF NOT EXISTS idx_subscribers_deleted_at ON subscribers(deleted_at);

-- +goose Down
-- Drop subscribers table
DROP TABLE IF EXISTS subscribers CASCADE;
