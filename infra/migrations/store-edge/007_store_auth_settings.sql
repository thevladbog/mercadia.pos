-- +goose Up
-- +goose StatementBegin
CREATE TABLE store_auth_settings (
    store_id TEXT PRIMARY KEY,
    failed_attempt_limit INTEGER NOT NULL,
    lockout_duration_seconds INTEGER NOT NULL,
    pos_auto_lock_seconds INTEGER NOT NULL,
    updated_by_id TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE auth_attempts (
    id TEXT PRIMARY KEY,
    store_id TEXT NOT NULL,
    actor_id TEXT NOT NULL,
    terminal_id TEXT NOT NULL DEFAULT '',
    credential_kind TEXT NOT NULL DEFAULT '',
    credential_fingerprint TEXT NOT NULL DEFAULT '',
    successful BOOLEAN NOT NULL,
    failure_reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_auth_attempts_store_actor_created ON auth_attempts (store_id, actor_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS auth_attempts;
DROP TABLE IF EXISTS store_auth_settings;
-- +goose StatementEnd
