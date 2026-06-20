-- +goose Up
CREATE TABLE color_schemes (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    logo_url TEXT NOT NULL DEFAULT '',
    accent_preset TEXT NOT NULL DEFAULT '',
    accent_color TEXT NOT NULL DEFAULT '',
    background_color TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft',
    store_ids TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS color_schemes;
