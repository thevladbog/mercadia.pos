-- +goose Up
ALTER TABLE fiscal_documents
    ADD COLUMN return_id TEXT REFERENCES returns (id);

CREATE UNIQUE INDEX idx_fiscal_documents_return_id
    ON fiscal_documents (return_id)
    WHERE return_id IS NOT NULL AND return_id <> '';

-- +goose Down
DROP INDEX IF EXISTS idx_fiscal_documents_return_id;

ALTER TABLE fiscal_documents
    DROP COLUMN IF EXISTS return_id;
