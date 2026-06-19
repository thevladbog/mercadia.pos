-- +goose Up
CREATE TABLE synced_fiscal_documents (
    store_id TEXT NOT NULL REFERENCES stores (id),
    id TEXT NOT NULL,
    receipt_id TEXT NOT NULL,
    kind TEXT NOT NULL,
    amount_minor BIGINT NOT NULL,
    device_id TEXT NOT NULL,
    fiscal_sign TEXT NOT NULL,
    fiscalized_at TIMESTAMPTZ NOT NULL,
    return_id TEXT,
    source_event_id TEXT NOT NULL,
    synced_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (store_id, id)
);

CREATE INDEX idx_synced_fiscal_documents_store_fiscalized_at ON synced_fiscal_documents (store_id, fiscalized_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_synced_fiscal_documents_store_fiscalized_at;
DROP TABLE IF EXISTS synced_fiscal_documents;
