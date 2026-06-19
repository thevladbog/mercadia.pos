-- +goose Up
CREATE TABLE synced_payments (
    store_id TEXT NOT NULL REFERENCES stores (id),
    id TEXT NOT NULL,
    receipt_id TEXT NOT NULL,
    method TEXT NOT NULL,
    amount_minor BIGINT NOT NULL,
    captured_at TIMESTAMPTZ NOT NULL,
    source_event_id TEXT NOT NULL,
    synced_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (store_id, id)
);

CREATE INDEX idx_synced_payments_store_captured_at ON synced_payments (store_id, captured_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_synced_payments_store_captured_at;
DROP TABLE IF EXISTS synced_payments;
