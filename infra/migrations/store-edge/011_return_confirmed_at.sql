-- +goose Up
-- +goose StatementBegin
ALTER TABLE returns ADD COLUMN confirmed_at TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE returns DROP COLUMN IF EXISTS confirmed_at;
-- +goose StatementEnd
