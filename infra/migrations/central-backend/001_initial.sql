-- +goose Up
CREATE TABLE stores (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    region TEXT NOT NULL DEFAULT 'default',
    registered_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE sync_events (
    id TEXT PRIMARY KEY,
    store_id TEXT NOT NULL REFERENCES stores (id),
    event_type TEXT NOT NULL,
    source_event_id TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    occurred_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL,
    UNIQUE (store_id, source_event_id)
);

CREATE TABLE catalog_products (
    store_id TEXT NOT NULL REFERENCES stores (id),
    id TEXT NOT NULL,
    name TEXT NOT NULL,
    barcodes TEXT[] NOT NULL DEFAULT '{}',
    unit_price_minor BIGINT NOT NULL,
    tax_category_id TEXT NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    version BIGINT NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (store_id, id)
);

CREATE TABLE central_users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE idempotency_records (
    operation TEXT NOT NULL,
    key TEXT NOT NULL,
    target_id TEXT NOT NULL DEFAULT '',
    fingerprint TEXT NOT NULL,
    result JSONB,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (operation, key)
);

CREATE INDEX idx_catalog_products_updated_at ON catalog_products (store_id, updated_at);
CREATE INDEX idx_sync_events_store_received_at ON sync_events (store_id, received_at);

-- +goose Down
DROP INDEX IF EXISTS idx_sync_events_store_received_at;
DROP INDEX IF EXISTS idx_catalog_products_updated_at;
DROP TABLE IF EXISTS idempotency_records;
DROP TABLE IF EXISTS central_users;
DROP TABLE IF EXISTS catalog_products;
DROP TABLE IF EXISTS sync_events;
DROP TABLE IF EXISTS stores;
