-- +goose Up
-- +goose StatementBegin
CREATE TABLE change_fund_requests (
    id TEXT PRIMARY KEY,
    store_id TEXT NOT NULL,
    shift_id TEXT NOT NULL,
    actor_id TEXT NOT NULL,
    amount_minor BIGINT NOT NULL,
    currency TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    fulfilled_by_id TEXT NOT NULL DEFAULT '',
    safe_id TEXT NOT NULL DEFAULT '',
    cash_movement_id TEXT NOT NULL DEFAULT '',
    denomination_breakdown JSONB,
    created_at TIMESTAMPTZ NOT NULL,
    fulfilled_at TIMESTAMPTZ
);

CREATE INDEX idx_change_fund_requests_store_id ON change_fund_requests (store_id);
CREATE INDEX idx_change_fund_requests_shift_id ON change_fund_requests (shift_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS change_fund_requests;
-- +goose StatementEnd
