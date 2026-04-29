-- +goose Up
-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id            SERIAL PRIMARY KEY,
    subscriber_id  INTEGER NOT NULL,
    transaction_id TEXT NOT NULL,
    type          TEXT NOT NULL,
    amount        DOUBLE PRECISION NOT NULL,
    currency      TEXT NOT NULL,
    status        TEXT NOT NULL,
    description   TEXT,
    parent_id     INTEGER,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_transactions_subscriber_id ON transactions(subscriber_id);
CREATE INDEX IF NOT EXISTS idx_transactions_transaction_id ON transactions(transaction_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);

-- +goose Down
-- Drop transactions table
DROP TABLE IF EXISTS transactions CASCADE;
