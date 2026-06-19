-- +goose Up
-- +goose StatementBegin
CREATE TABLE store_actors (
    id TEXT PRIMARY KEY,
    pin TEXT NOT NULL,
    roles JSONB NOT NULL DEFAULT '[]'::jsonb
);

CREATE TABLE sessions (
    token TEXT PRIMARY KEY,
    actor_id TEXT NOT NULL REFERENCES store_actors (id),
    roles JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_sessions_actor_id ON sessions (actor_id);
CREATE INDEX idx_sessions_expires_at ON sessions (expires_at);

CREATE TABLE returns (
    id TEXT PRIMARY KEY,
    store_id TEXT NOT NULL,
    receipt_id TEXT NOT NULL DEFAULT '',
    kind TEXT NOT NULL,
    lines JSONB NOT NULL DEFAULT '[]'::jsonb,
    reason TEXT NOT NULL,
    actor_id TEXT NOT NULL,
    approved_by_id TEXT NOT NULL DEFAULT '',
    total_minor BIGINT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_returns_store_id ON returns (store_id);
CREATE INDEX idx_returns_receipt_id ON returns (receipt_id);

CREATE TABLE operation_journal_entries (
    id TEXT PRIMARY KEY,
    store_id TEXT NOT NULL,
    operation_type TEXT NOT NULL,
    actor_id TEXT NOT NULL,
    reference_id TEXT NOT NULL DEFAULT '',
    summary TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_operation_journal_store_created ON operation_journal_entries (store_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS operation_journal_entries;
DROP TABLE IF EXISTS returns;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS store_actors;
-- +goose StatementEnd
