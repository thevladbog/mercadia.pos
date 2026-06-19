package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	platformmigrate "mercadia.dev/pos/platform/migrate"
	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/domain"
)

type Store struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	if err := RunMigrations(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	dir := migrationsDir()
	if dir == "" {
		return fmt.Errorf("migrations directory not found")
	}

	result, err := platformmigrate.UpPool(ctx, pool, "central-backend", dir)
	if err != nil {
		platformmigrate.LogError("central-backend", dir, err)
		return err
	}
	platformmigrate.LogResult(result)
	return nil
}

func migrationsDir() string {
	return platformmigrate.FindMigrationsDir(
		"MERCADIA_CENTRAL_BACKEND_MIGRATIONS_DIR",
		"infra/migrations/central-backend",
	)
}

func (s *Store) SaveStore(ctx context.Context, store domain.Store) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO stores (id, name, region, registered_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			region = EXCLUDED.region,
			updated_at = EXCLUDED.updated_at
	`, store.ID, store.Name, store.Region, store.RegisteredAt, store.UpdatedAt)
	return err
}

func (s *Store) FindStore(ctx context.Context, storeID string) (domain.Store, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, region, registered_at, updated_at
		FROM stores
		WHERE id = $1
	`, storeID)
	return scanStore(row)
}

func (s *Store) ListStores(ctx context.Context) ([]domain.Store, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, region, registered_at, updated_at
		FROM stores
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stores := []domain.Store{}
	for rows.Next() {
		store, err := scanStore(rows)
		if err != nil {
			return nil, err
		}
		stores = append(stores, store)
	}
	return stores, rows.Err()
}

func (s *Store) CountStores(ctx context.Context) (int, error) {
	row := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM stores`)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) SaveSyncEvent(ctx context.Context, event domain.SyncEvent) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sync_events (id, store_id, event_type, source_event_id, payload, occurred_at, received_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, event.ID, event.StoreID, event.EventType, event.SourceEventID, []byte(event.Payload), event.OccurredAt, event.ReceivedAt)
	if err != nil && isUniqueViolation(err) {
		return app.ErrSyncEventDuplicate
	}
	return err
}

func (s *Store) ExistsSyncEvent(ctx context.Context, storeID string, sourceEventID string) (bool, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM sync_events WHERE store_id = $1 AND source_event_id = $2
		)
	`, storeID, sourceEventID)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Store) SaveProduct(ctx context.Context, product domain.CatalogProduct) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO catalog_products (
			store_id, id, name, barcodes, unit_price_minor, tax_category_id, active, version, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (store_id, id) DO UPDATE SET
			name = EXCLUDED.name,
			barcodes = EXCLUDED.barcodes,
			unit_price_minor = EXCLUDED.unit_price_minor,
			tax_category_id = EXCLUDED.tax_category_id,
			active = EXCLUDED.active,
			version = EXCLUDED.version,
			updated_at = EXCLUDED.updated_at
	`, product.StoreID, product.ID, product.Name, product.Barcodes, product.UnitPriceMinor, product.TaxCategoryID, product.Active, product.Version, product.UpdatedAt)
	return err
}

func (s *Store) FindProduct(ctx context.Context, storeID string, productID string) (domain.CatalogProduct, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT store_id, id, name, barcodes, unit_price_minor, tax_category_id, active, version, updated_at
		FROM catalog_products
		WHERE store_id = $1 AND id = $2
	`, storeID, productID)
	return scanProduct(row)
}

func (s *Store) ListProducts(ctx context.Context, storeID string) ([]domain.CatalogProduct, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT store_id, id, name, barcodes, unit_price_minor, tax_category_id, active, version, updated_at
		FROM catalog_products
		WHERE store_id = $1
		ORDER BY id
	`, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProducts(rows)
}

func (s *Store) ListProductsSince(ctx context.Context, storeID string, since time.Time) ([]domain.CatalogProduct, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT store_id, id, name, barcodes, unit_price_minor, tax_category_id, active, version, updated_at
		FROM catalog_products
		WHERE store_id = $1 AND updated_at > $2
		ORDER BY id
	`, storeID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProducts(rows)
}

func (s *Store) Find(ctx context.Context, operation string, key string) (app.IdempotencyRecord, bool, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT operation, key, target_id, fingerprint, result, created_at
		FROM idempotency_records
		WHERE operation = $1 AND key = $2
	`, operation, key)

	var record app.IdempotencyRecord
	var resultJSON []byte
	if err := row.Scan(&record.Operation, &record.Key, &record.TargetID, &record.Fingerprint, &resultJSON, &record.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return app.IdempotencyRecord{}, false, nil
		}
		return app.IdempotencyRecord{}, false, err
	}
	if len(resultJSON) > 0 {
		result, err := decodeIdempotencyResult(operation, resultJSON)
		if err != nil {
			return app.IdempotencyRecord{}, false, err
		}
		record.Result = result
	}
	return record, true, nil
}

func (s *Store) Save(ctx context.Context, record app.IdempotencyRecord) error {
	resultJSON, err := json.Marshal(record.Result)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO idempotency_records (operation, key, target_id, fingerprint, result, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (operation, key) DO UPDATE SET
			target_id = EXCLUDED.target_id,
			fingerprint = EXCLUDED.fingerprint,
			result = EXCLUDED.result,
			created_at = EXCLUDED.created_at
	`, record.Operation, record.Key, record.TargetID, record.Fingerprint, resultJSON, record.CreatedAt)
	return err
}

func decodeIdempotencyResult(operation string, data []byte) (any, error) {
	switch operation {
	case "register_store":
		var result app.StoreResult
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return result, nil
	case "accept_sync_events":
		var result app.SyncEventsResult
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return result, nil
	default:
		var result any
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return result, nil
	}
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanStore(row rowScanner) (domain.Store, error) {
	var store domain.Store
	if err := row.Scan(&store.ID, &store.Name, &store.Region, &store.RegisteredAt, &store.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Store{}, app.ErrStoreNotFound
		}
		return domain.Store{}, err
	}
	return store, nil
}

func scanProduct(row rowScanner) (domain.CatalogProduct, error) {
	var product domain.CatalogProduct
	if err := row.Scan(&product.StoreID, &product.ID, &product.Name, &product.Barcodes, &product.UnitPriceMinor, &product.TaxCategoryID, &product.Active, &product.Version, &product.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CatalogProduct{}, app.ErrCatalogProductNotFound
		}
		return domain.CatalogProduct{}, err
	}
	return product, nil
}

func scanProducts(rows pgx.Rows) ([]domain.CatalogProduct, error) {
	products := []domain.CatalogProduct{}
	for rows.Next() {
		var product domain.CatalogProduct
		if err := rows.Scan(&product.StoreID, &product.ID, &product.Name, &product.Barcodes, &product.UnitPriceMinor, &product.TaxCategoryID, &product.Active, &product.Version, &product.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(products, func(i, j int) bool {
		return products[i].ID < products[j].ID
	})
	return products, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
