-- +goose Up
CREATE TABLE synced_returns (
    store_id TEXT NOT NULL REFERENCES stores (id),
    id TEXT NOT NULL,
    receipt_id TEXT NOT NULL,
    total_minor BIGINT NOT NULL,
    payment_ids JSONB NOT NULL DEFAULT '[]',
    cash_movement_id TEXT,
    actor_id TEXT NOT NULL,
    settled_at TIMESTAMPTZ NOT NULL,
    source_event_id TEXT NOT NULL,
    synced_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (store_id, id)
);

CREATE INDEX idx_synced_returns_store_settled_at ON synced_returns (store_id, settled_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_synced_returns_store_settled_at;
DROP TABLE IF EXISTS synced_returns;
