-- +goose Up
-- Create esim_profiles table
CREATE TABLE IF NOT EXISTS esim_profiles (
    iccid          TEXT PRIMARY KEY,
    eid            TEXT,
    imsi           TEXT,
    mcc            TEXT,
    mnc            TEXT,
    profile_type   TEXT,
    state          TEXT,
    tenant_id      TEXT,
    activation_code TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_esim_profiles_eid ON esim_profiles(eid);
CREATE INDEX IF NOT EXISTS idx_esim_profiles_imsi ON esim_profiles(imsi);
CREATE INDEX IF NOT EXISTS idx_esim_profiles_state ON esim_profiles(state);
CREATE INDEX IF NOT EXISTS idx_esim_profiles_tenant_id ON esim_profiles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_esim_profiles_created_at ON esim_profiles(created_at);

-- +goose Down
-- Drop esim_profiles table
DROP TABLE IF EXISTS esim_profiles CASCADE;
