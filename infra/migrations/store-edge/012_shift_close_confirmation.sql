-- +goose Up
-- +goose StatementBegin
ALTER TABLE shifts ADD COLUMN closing_actor_id TEXT NOT NULL DEFAULT '';
ALTER TABLE shifts ADD COLUMN closing_safe_id TEXT NOT NULL DEFAULT '';
ALTER TABLE shifts ADD COLUMN closing_approved_by_id TEXT NOT NULL DEFAULT '';
ALTER TABLE shifts ADD COLUMN awaiting_confirmation_since TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE shifts DROP COLUMN IF EXISTS awaiting_confirmation_since;
ALTER TABLE shifts DROP COLUMN IF EXISTS closing_approved_by_id;
ALTER TABLE shifts DROP COLUMN IF EXISTS closing_safe_id;
ALTER TABLE shifts DROP COLUMN IF EXISTS closing_actor_id;
-- +goose StatementEnd
