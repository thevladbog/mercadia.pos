-- +goose Up
-- +goose StatementBegin
ALTER TABLE store_actors
    ADD COLUMN credential_policy JSONB,
    ADD COLUMN credential_bindings JSONB NOT NULL DEFAULT '[]'::jsonb;

CREATE TABLE store_credential_policies (
    store_id TEXT PRIMARY KEY,
    policy JSONB NOT NULL
);

ALTER TABLE sessions
    ADD COLUMN credential_factor JSONB;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sessions DROP COLUMN IF EXISTS credential_factor;
DROP TABLE IF EXISTS store_credential_policies;
ALTER TABLE store_actors
    DROP COLUMN IF EXISTS credential_bindings,
    DROP COLUMN IF EXISTS credential_policy;
-- +goose StatementEnd
