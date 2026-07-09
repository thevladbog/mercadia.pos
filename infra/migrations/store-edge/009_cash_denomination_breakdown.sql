-- +goose Up
-- +goose StatementBegin
ALTER TABLE cash_movements ADD COLUMN denomination_breakdown JSONB;
ALTER TABLE cash_recounts ADD COLUMN denomination_breakdown JSONB;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE cash_recounts DROP COLUMN IF EXISTS denomination_breakdown;
ALTER TABLE cash_movements DROP COLUMN IF EXISTS denomination_breakdown;
-- +goose StatementEnd
