-- +goose Up
CREATE TABLE layout_templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    kind TEXT NOT NULL,
    accent_preset TEXT NOT NULL DEFAULT '',
    accent_color TEXT NOT NULL DEFAULT '',
    color_scheme_id TEXT REFERENCES color_schemes (id),
    grid JSONB NOT NULL DEFAULT '{}',
    store_id TEXT REFERENCES stores (id),
    terminal_type TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft',
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS layout_templates;
