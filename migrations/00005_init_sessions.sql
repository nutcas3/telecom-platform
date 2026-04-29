-- +goose Up
-- Create sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id           SERIAL PRIMARY KEY,
    subscriber_id INTEGER NOT NULL,
    session_id   TEXT NOT NULL UNIQUE,
    pdu_address  TEXT,
    dnn          TEXT,
    snssai_sst   SMALLINT,
    snssai_sd    INTEGER,
    qos_qci      SMALLINT,
    qos_arpa     INTEGER,
    qos_mbr_uplink   INTEGER,
    qos_mbr_downlink INTEGER,
    qos_gbr_uplink   INTEGER,
    qos_gbr_downlink INTEGER,
    qos_priority SMALLINT,
    status       TEXT NOT NULL DEFAULT 'active',
    start_time   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_time     TIMESTAMPTZ,
    data_used    BIGINT NOT NULL DEFAULT 0,
    voice_used   INTEGER NOT NULL DEFAULT 0,
    sms_used     INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_sessions_subscriber_id ON sessions(subscriber_id);
CREATE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_deleted_at ON sessions(deleted_at);

-- +goose Down
-- Drop sessions table
DROP TABLE IF EXISTS sessions CASCADE;
