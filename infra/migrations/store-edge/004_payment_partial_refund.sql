-- +goose Up
ALTER TABLE payments
    ADD COLUMN refunded_amount_minor BIGINT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE payments
    DROP COLUMN IF EXISTS refunded_amount_minor;
