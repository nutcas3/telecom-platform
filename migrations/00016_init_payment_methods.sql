-- +goose Up
-- Create payment_methods table
CREATE TABLE IF NOT EXISTS payment_methods (
    id            TEXT PRIMARY KEY,
    subscriber_id INTEGER NOT NULL,
    gateway_id    TEXT,
    type          TEXT NOT NULL,
    customer_id   TEXT,
    last4         TEXT,
    brand         TEXT,
    expiry_month  INTEGER,
    expiry_year   INTEGER,
    is_default    BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_payment_methods_subscriber_id ON payment_methods(subscriber_id);
CREATE INDEX IF NOT EXISTS idx_payment_methods_gateway_id ON payment_methods(gateway_id);

-- +goose Down
-- Drop payment_methods table
DROP TABLE IF EXISTS payment_methods CASCADE;
