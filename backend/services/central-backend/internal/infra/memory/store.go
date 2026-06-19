package memory

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/domain"
)

type Store struct {
	mu          sync.RWMutex
	stores      map[string]domain.Store
	syncEvents  map[string]domain.SyncEvent
	syncByStore map[string]map[string]string
	products    map[string]domain.CatalogProduct
	users       map[string]domain.CentralUser
	idempotency map[string]app.IdempotencyRecord
}

type StoreOption func(*Store)

func NewStore(options ...StoreOption) *Store {
	store := &Store{
		stores:      map[string]domain.Store{},
		syncEvents:  map[string]domain.SyncEvent{},
		syncByStore: map[string]map[string]string{},
		products:    map[string]domain.CatalogProduct{},
		users:       map[string]domain.CentralUser{},
		idempotency: map[string]app.IdempotencyRecord{},
	}
	for _, option := range options {
		option(store)
	}
	return store
}

func WithStores(stores ...domain.Store) StoreOption {
	return func(store *Store) {
		for _, item := range stores {
			store.stores[item.ID] = item
		}
	}
}

func WithProducts(products ...domain.CatalogProduct) StoreOption {
	return func(store *Store) {
		for _, product := range products {
			store.products[productKey(product.StoreID, product.ID)] = cloneProduct(product)
		}
	}
}

func (s *Store) SaveStore(ctx context.Context, store domain.Store) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stores[store.ID] = store
	return nil
}

func (s *Store) FindStore(ctx context.Context, storeID string) (domain.Store, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	store, ok := s.stores[storeID]
	if !ok {
		return domain.Store{}, app.ErrStoreNotFound
	}
	return store, nil
}

func (s *Store) ListStores(ctx context.Context) ([]domain.Store, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stores := make([]domain.Store, 0, len(s.stores))
	for _, store := range s.stores {
		stores = append(stores, store)
	}
	sort.Slice(stores, func(i, j int) bool {
		return stores[i].ID < stores[j].ID
	})
	return stores, nil
}

func (s *Store) CountStores(ctx context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.stores), nil
}

func (s *Store) SaveSyncEvent(ctx context.Context, event domain.SyncEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.syncByStore[event.StoreID] == nil {
		s.syncByStore[event.StoreID] = map[string]string{}
	}
	if existingID, ok := s.syncByStore[event.StoreID][event.SourceEventID]; ok {
		_ = existingID
		return app.ErrSyncEventDuplicate
	}
	s.syncEvents[event.ID] = cloneSyncEvent(event)
	s.syncByStore[event.StoreID][event.SourceEventID] = event.ID
	return nil
}

func (s *Store) ExistsSyncEvent(ctx context.Context, storeID string, sourceEventID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	byStore, ok := s.syncByStore[storeID]
	if !ok {
		return false, nil
	}
	_, ok = byStore[sourceEventID]
	return ok, nil
}

func (s *Store) SaveProduct(ctx context.Context, product domain.CatalogProduct) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.products[productKey(product.StoreID, product.ID)] = cloneProduct(product)
	return nil
}

func (s *Store) FindProduct(ctx context.Context, storeID string, productID string) (domain.CatalogProduct, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	product, ok := s.products[productKey(storeID, productID)]
	if !ok {
		return domain.CatalogProduct{}, app.ErrCatalogProductNotFound
	}
	return cloneProduct(product), nil
}

func (s *Store) ListProducts(ctx context.Context, storeID string) ([]domain.CatalogProduct, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	products := []domain.CatalogProduct{}
	for _, product := range s.products {
		if product.StoreID == storeID {
			products = append(products, cloneProduct(product))
		}
	}
	sort.Slice(products, func(i, j int) bool {
		return products[i].ID < products[j].ID
	})
	return products, nil
}

func (s *Store) ListProductsSince(ctx context.Context, storeID string, since time.Time) ([]domain.CatalogProduct, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	products := []domain.CatalogProduct{}
	for _, product := range s.products {
		if product.StoreID == storeID && product.UpdatedAt.After(since) {
			products = append(products, cloneProduct(product))
		}
	}
	sort.Slice(products, func(i, j int) bool {
		return products[i].ID < products[j].ID
	})
	return products, nil
}

func (s *Store) Find(ctx context.Context, operation string, key string) (app.IdempotencyRecord, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.idempotency[idempotencyMapKey(operation, key)]
	return record, ok, nil
}

func (s *Store) Save(ctx context.Context, record app.IdempotencyRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.idempotency[idempotencyMapKey(record.Operation, record.Key)] = record
	return nil
}

func productKey(storeID string, productID string) string {
	return storeID + "\x00" + productID
}

func idempotencyMapKey(operation string, key string) string {
	return operation + "\x00" + key
}

func cloneProduct(product domain.CatalogProduct) domain.CatalogProduct {
	product.Barcodes = append([]string(nil), product.Barcodes...)
	return product
}

func cloneSyncEvent(event domain.SyncEvent) domain.SyncEvent {
	if len(event.Payload) > 0 {
		event.Payload = append(json.RawMessage(nil), event.Payload...)
	}
	return event
}
