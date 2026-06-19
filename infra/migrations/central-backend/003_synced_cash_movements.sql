-- +goose Up
CREATE TABLE synced_cash_movements (
    store_id TEXT NOT NULL REFERENCES stores (id),
    id TEXT NOT NULL,
    type TEXT NOT NULL,
    from_container_id TEXT NOT NULL,
    from_container_type TEXT NOT NULL,
    to_container_id TEXT NOT NULL,
    to_container_type TEXT NOT NULL,
    amount_minor BIGINT NOT NULL,
    currency TEXT NOT NULL DEFAULT '',
    actor_id TEXT NOT NULL,
    posted_at TIMESTAMPTZ NOT NULL,
    source_event_id TEXT NOT NULL,
    synced_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (store_id, id)
);

CREATE INDEX idx_synced_cash_movements_store_posted_at ON synced_cash_movements (store_id, posted_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_synced_cash_movements_store_posted_at;
DROP TABLE IF EXISTS synced_cash_movements;
