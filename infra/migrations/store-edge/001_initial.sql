-- +goose Up
-- +goose StatementBegin
CREATE TABLE idempotency_records (
    operation TEXT NOT NULL,
    key TEXT NOT NULL,
    target_id TEXT NOT NULL DEFAULT '',
    fingerprint TEXT NOT NULL DEFAULT '',
    result JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (operation, key)
);

CREATE TABLE operational_days (
    id TEXT PRIMARY KEY,
    store_id TEXT NOT NULL,
    business_date TEXT NOT NULL,
    status TEXT NOT NULL,
    opened_by_id TEXT NOT NULL,
    closed_by_id TEXT NOT NULL DEFAULT '',
    opened_at TIMESTAMPTZ NOT NULL,
    closed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_operational_days_store_status ON operational_days (store_id, status);

CREATE TABLE shifts (
    id TEXT PRIMARY KEY,
    store_id TEXT NOT NULL,
    operational_day_id TEXT NOT NULL DEFAULT '',
    business_date TEXT NOT NULL DEFAULT '',
    terminal_id TEXT NOT NULL,
    cashier_id TEXT NOT NULL,
    drawer_id TEXT NOT NULL,
    status TEXT NOT NULL,
    opening_cash_minor BIGINT NOT NULL DEFAULT 0,
    closing_cash_minor BIGINT NOT NULL DEFAULT 0,
    opened_at TIMESTAMPTZ NOT NULL,
    closed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_shifts_store_status ON shifts (store_id, status);
CREATE INDEX idx_shifts_operational_day_id ON shifts (operational_day_id);
CREATE INDEX idx_shifts_terminal_open ON shifts (terminal_id) WHERE status = 'open';
CREATE INDEX idx_shifts_cashier_open ON shifts (cashier_id) WHERE status = 'open';

CREATE TABLE receipts (
    id TEXT PRIMARY KEY,
    store_id TEXT NOT NULL,
    operational_day_id TEXT NOT NULL DEFAULT '',
    business_date TEXT NOT NULL DEFAULT '',
    shift_id TEXT NOT NULL DEFAULT '',
    terminal_id TEXT NOT NULL,
    cashier_id TEXT NOT NULL,
    drawer_id TEXT NOT NULL DEFAULT '',
    channel TEXT NOT NULL,
    status TEXT NOT NULL,
    lines JSONB NOT NULL DEFAULT '[]'::jsonb,
    cancel_reason TEXT NOT NULL DEFAULT '',
    cancelled_by_id TEXT NOT NULL DEFAULT '',
    cancel_approved_by_id TEXT NOT NULL DEFAULT '',
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_receipts_shift_id ON receipts (shift_id);
CREATE INDEX idx_receipts_operational_day_id ON receipts (operational_day_id);
CREATE INDEX idx_receipts_store_business_date ON receipts (store_id, business_date);
CREATE INDEX idx_receipts_store_status ON receipts (store_id, status);

CREATE TABLE payments (
    id TEXT PRIMARY KEY,
    receipt_id TEXT NOT NULL REFERENCES receipts (id),
    method TEXT NOT NULL,
    status TEXT NOT NULL,
    amount_minor BIGINT NOT NULL,
    provider_reference TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    captured_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_payments_receipt_id ON payments (receipt_id);

CREATE TABLE fiscal_documents (
    id TEXT PRIMARY KEY,
    receipt_id TEXT NOT NULL REFERENCES receipts (id),
    kind TEXT NOT NULL,
    status TEXT NOT NULL,
    amount_minor BIGINT NOT NULL,
    device_id TEXT NOT NULL,
    fiscal_sign TEXT NOT NULL,
    fiscalized_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_fiscal_documents_receipt_id ON fiscal_documents (receipt_id);

CREATE TABLE cash_movements (
    id TEXT PRIMARY KEY,
    store_id TEXT NOT NULL,
    type TEXT NOT NULL,
    from_container_id TEXT NOT NULL,
    from_container_type TEXT NOT NULL,
    to_container_id TEXT NOT NULL,
    to_container_type TEXT NOT NULL,
    amount_minor BIGINT NOT NULL,
    currency TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    actor_id TEXT NOT NULL,
    approved_by_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_cash_movements_store_id ON cash_movements (store_id);

CREATE TABLE cash_recounts (
    id TEXT PRIMARY KEY,
    store_id TEXT NOT NULL,
    business_date TEXT NOT NULL DEFAULT '',
    container_id TEXT NOT NULL,
    container_type TEXT NOT NULL,
    currency TEXT NOT NULL,
    expected_minor BIGINT NOT NULL,
    counted_minor BIGINT NOT NULL,
    discrepancy_minor BIGINT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    actor_id TEXT NOT NULL,
    approved_by_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    resolution_status TEXT NOT NULL,
    resolution_note TEXT NOT NULL DEFAULT '',
    resolved_by_id TEXT NOT NULL DEFAULT '',
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_cash_recounts_store_id ON cash_recounts (store_id);
CREATE INDEX idx_cash_recounts_store_business_date ON cash_recounts (store_id, business_date);

CREATE TABLE terminals (
    id TEXT PRIMARY KEY,
    store_id TEXT NOT NULL,
    kind TEXT NOT NULL,
    status TEXT NOT NULL,
    software_version TEXT NOT NULL DEFAULT '',
    last_seen_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_terminals_store_id ON terminals (store_id);

CREATE TABLE products (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    barcodes JSONB NOT NULL DEFAULT '[]'::jsonb,
    unit_price_minor BIGINT NOT NULL,
    tax_category_id TEXT NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_products_barcodes ON products USING GIN (barcodes);

CREATE TABLE outbox_events (
    id TEXT PRIMARY KEY,
    aggregate_type TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    published_at TIMESTAMPTZ
);

CREATE INDEX idx_outbox_events_unpublished ON outbox_events (created_at) WHERE published_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS terminals;
DROP TABLE IF EXISTS cash_recounts;
DROP TABLE IF EXISTS cash_movements;
DROP TABLE IF EXISTS fiscal_documents;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS receipts;
DROP TABLE IF EXISTS shifts;
DROP TABLE IF EXISTS operational_days;
DROP TABLE IF EXISTS idempotency_records;
-- +goose StatementEnd
