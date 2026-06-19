-- +goose Up
ALTER TABLE synced_payments
    ADD COLUMN status TEXT NOT NULL DEFAULT 'captured',
    ADD COLUMN cancelled_at TIMESTAMPTZ,
    ADD COLUMN refunded_amount_minor BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN remaining_amount_minor BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN last_event_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE synced_payments
SET last_event_id = source_event_id,
    updated_at = synced_at
WHERE last_event_id = '';

-- +goose Down
ALTER TABLE synced_payments
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS last_event_id,
    DROP COLUMN IF EXISTS remaining_amount_minor,
    DROP COLUMN IF EXISTS refunded_amount_minor,
    DROP COLUMN IF EXISTS cancelled_at,
    DROP COLUMN IF EXISTS status;
