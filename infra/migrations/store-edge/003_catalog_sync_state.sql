-- +goose Up
CREATE TABLE IF NOT EXISTS catalog_sync_state (
    store_id TEXT PRIMARY KEY,
    last_synced_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS catalog_sync_state;
