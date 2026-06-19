-- +goose Up
ALTER TABLE central_users
    ADD COLUMN password_hash TEXT NOT NULL DEFAULT '',
    ADD COLUMN roles JSONB NOT NULL DEFAULT '[]'::jsonb;

CREATE TABLE central_sessions (
    token TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES central_users (id),
    roles JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_central_sessions_user_id ON central_sessions (user_id);
CREATE INDEX idx_central_sessions_expires_at ON central_sessions (expires_at);

-- +goose Down
DROP INDEX IF EXISTS idx_central_sessions_expires_at;
DROP INDEX IF EXISTS idx_central_sessions_user_id;
DROP TABLE IF EXISTS central_sessions;
ALTER TABLE central_users DROP COLUMN IF EXISTS roles;
ALTER TABLE central_users DROP COLUMN IF EXISTS password_hash;
