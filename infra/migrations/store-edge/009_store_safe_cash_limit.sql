-- +goose Up
-- +goose StatementBegin
ALTER TABLE store_auth_settings ADD COLUMN safe_cash_limit_minor BIGINT NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE store_auth_settings DROP COLUMN IF EXISTS safe_cash_limit_minor;
-- +goose StatementEnd
