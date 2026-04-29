-- +goose Up
-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id           SERIAL PRIMARY KEY,
    subscriber_id INTEGER NOT NULL,
    type         TEXT NOT NULL,
    title        TEXT NOT NULL,
    message      TEXT NOT NULL,
    read         BOOLEAN NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id           SERIAL PRIMARY KEY,
    type         TEXT NOT NULL,
    severity     TEXT NOT NULL,
    message      TEXT NOT NULL,
    subscriber_id INTEGER,
    timestamp    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved     BOOLEAN NOT NULL DEFAULT false
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_notifications_subscriber_id ON notifications(subscriber_id);
CREATE INDEX IF NOT EXISTS idx_notifications_read ON notifications(read);
CREATE INDEX IF NOT EXISTS idx_alerts_subscriber_id ON alerts(subscriber_id);
CREATE INDEX IF NOT EXISTS idx_alerts_type ON alerts(type);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_resolved ON alerts(resolved);

-- +goose Down
-- Drop notifications and alerts tables
DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS alerts CASCADE;
