-- +goose Up
-- +goose StatementBegin
-- Staff PINs were previously stored as plaintext in store_actors.pin. This
-- renames the column to pin_hash to hold bcrypt hashes instead. Actors are
-- currently seed-only demo data (see Store Edge's SeedDemoActors), and the
-- startup seed upsert overwrites every renamed row with a freshly computed
-- bcrypt hash of the same demo PIN, so no data-migration step is needed here.
-- If actors ever become creatable/updatable outside the seed path, a real
-- PIN-reset flow will be required for any row not covered by the seed upsert.
ALTER TABLE store_actors RENAME COLUMN pin TO pin_hash;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE store_actors RENAME COLUMN pin_hash TO pin;
-- +goose StatementEnd
