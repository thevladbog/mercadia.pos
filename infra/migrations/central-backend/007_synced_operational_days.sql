-- +goose Up
CREATE TABLE synced_operational_days (
    store_id TEXT NOT NULL REFERENCES stores (id),
    id TEXT NOT NULL,
    business_date TEXT NOT NULL,
    closed_by_id TEXT NOT NULL,
    closed_at TIMESTAMPTZ NOT NULL,
    source_event_id TEXT NOT NULL,
    synced_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (store_id, id)
);

CREATE INDEX idx_synced_operational_days_store_closed_at ON synced_operational_days (store_id, closed_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_synced_operational_days_store_closed_at;
DROP TABLE IF EXISTS synced_operational_days;
