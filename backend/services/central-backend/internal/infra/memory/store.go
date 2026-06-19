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
	mu              sync.RWMutex
	stores          map[string]domain.Store
	syncEvents      map[string]domain.SyncEvent
	syncByStore     map[string]map[string]string
	products        map[string]domain.CatalogProduct
	payments        map[string]domain.SyncedPayment
	cashMovements   map[string]domain.SyncedCashMovement
	fiscalDocuments map[string]domain.SyncedFiscalDocument
	returns         map[string]domain.SyncedReturn
	operationalDays map[string]domain.SyncedOperationalDay
	users           map[string]domain.CentralUser
	sessions        map[string]domain.CentralSession
	idempotency     map[string]app.IdempotencyRecord
}

type StoreOption func(*Store)

func NewStore(options ...StoreOption) *Store {
	store := &Store{
		stores:          map[string]domain.Store{},
		syncEvents:      map[string]domain.SyncEvent{},
		syncByStore:     map[string]map[string]string{},
		products:        map[string]domain.CatalogProduct{},
		payments:        map[string]domain.SyncedPayment{},
		cashMovements:   map[string]domain.SyncedCashMovement{},
		fiscalDocuments: map[string]domain.SyncedFiscalDocument{},
		returns:         map[string]domain.SyncedReturn{},
		operationalDays: map[string]domain.SyncedOperationalDay{},
		users:           map[string]domain.CentralUser{},
		sessions:        map[string]domain.CentralSession{},
		idempotency:     map[string]app.IdempotencyRecord{},
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

func (s *Store) ListSyncEvents(ctx context.Context, storeID string, limit, offset int) ([]domain.SyncEvent, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]domain.SyncEvent, 0)
	for _, event := range s.syncEvents {
		if event.StoreID == storeID {
			events = append(events, cloneSyncEvent(event))
		}
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].ReceivedAt.Equal(events[j].ReceivedAt) {
			return events[i].ID > events[j].ID
		}
		return events[i].ReceivedAt.After(events[j].ReceivedAt)
	})

	total := len(events)
	if offset >= total {
		return []domain.SyncEvent{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return append([]domain.SyncEvent(nil), events[offset:end]...), total, nil
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

func (s *Store) SavePayment(ctx context.Context, payment domain.SyncedPayment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.payments[productKey(payment.StoreID, payment.ID)] = cloneSyncedPayment(payment)
	return nil
}

func (s *Store) FindPayment(ctx context.Context, storeID string, paymentID string) (domain.SyncedPayment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	payment, ok := s.payments[productKey(storeID, paymentID)]
	if !ok {
		return domain.SyncedPayment{}, app.ErrPaymentNotFound
	}
	return cloneSyncedPayment(payment), nil
}

func (s *Store) ListPayments(ctx context.Context, storeID string, limit, offset int) ([]domain.SyncedPayment, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	payments := make([]domain.SyncedPayment, 0)
	for _, payment := range s.payments {
		if payment.StoreID == storeID {
			payments = append(payments, cloneSyncedPayment(payment))
		}
	}
	sort.Slice(payments, func(i, j int) bool {
		if payments[i].CapturedAt.Equal(payments[j].CapturedAt) {
			return payments[i].ID > payments[j].ID
		}
		return payments[i].CapturedAt.After(payments[j].CapturedAt)
	})

	total := len(payments)
	if offset >= total {
		return []domain.SyncedPayment{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return append([]domain.SyncedPayment(nil), payments[offset:end]...), total, nil
}

func (s *Store) SaveCashMovement(ctx context.Context, movement domain.SyncedCashMovement) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cashMovements[productKey(movement.StoreID, movement.ID)] = cloneSyncedCashMovement(movement)
	return nil
}

func (s *Store) FindCashMovement(ctx context.Context, storeID string, cashMovementID string) (domain.SyncedCashMovement, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	movement, ok := s.cashMovements[productKey(storeID, cashMovementID)]
	if !ok {
		return domain.SyncedCashMovement{}, app.ErrCashMovementNotFound
	}
	return cloneSyncedCashMovement(movement), nil
}

func (s *Store) ListCashMovements(ctx context.Context, storeID string, limit, offset int) ([]domain.SyncedCashMovement, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	movements := make([]domain.SyncedCashMovement, 0)
	for _, movement := range s.cashMovements {
		if movement.StoreID == storeID {
			movements = append(movements, cloneSyncedCashMovement(movement))
		}
	}
	sort.Slice(movements, func(i, j int) bool {
		if movements[i].PostedAt.Equal(movements[j].PostedAt) {
			return movements[i].ID > movements[j].ID
		}
		return movements[i].PostedAt.After(movements[j].PostedAt)
	})

	total := len(movements)
	if offset >= total {
		return []domain.SyncedCashMovement{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return append([]domain.SyncedCashMovement(nil), movements[offset:end]...), total, nil
}

func (s *Store) SaveFiscalDocument(ctx context.Context, document domain.SyncedFiscalDocument) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.fiscalDocuments[productKey(document.StoreID, document.ID)] = cloneSyncedFiscalDocument(document)
	return nil
}

func (s *Store) FindFiscalDocument(ctx context.Context, storeID string, fiscalDocumentID string) (domain.SyncedFiscalDocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	document, ok := s.fiscalDocuments[productKey(storeID, fiscalDocumentID)]
	if !ok {
		return domain.SyncedFiscalDocument{}, app.ErrFiscalDocumentNotFound
	}
	return cloneSyncedFiscalDocument(document), nil
}

func (s *Store) ListFiscalDocuments(ctx context.Context, storeID string, limit, offset int) ([]domain.SyncedFiscalDocument, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	documents := make([]domain.SyncedFiscalDocument, 0)
	for _, document := range s.fiscalDocuments {
		if document.StoreID == storeID {
			documents = append(documents, cloneSyncedFiscalDocument(document))
		}
	}
	sort.Slice(documents, func(i, j int) bool {
		if documents[i].FiscalizedAt.Equal(documents[j].FiscalizedAt) {
			return documents[i].ID > documents[j].ID
		}
		return documents[i].FiscalizedAt.After(documents[j].FiscalizedAt)
	})

	total := len(documents)
	if offset >= total {
		return []domain.SyncedFiscalDocument{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return append([]domain.SyncedFiscalDocument(nil), documents[offset:end]...), total, nil
}

func (s *Store) SaveReturn(ctx context.Context, ret domain.SyncedReturn) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.returns[productKey(ret.StoreID, ret.ID)] = cloneSyncedReturn(ret)
	return nil
}

func (s *Store) FindReturn(ctx context.Context, storeID string, returnID string) (domain.SyncedReturn, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ret, ok := s.returns[productKey(storeID, returnID)]
	if !ok {
		return domain.SyncedReturn{}, app.ErrReturnNotFound
	}
	return cloneSyncedReturn(ret), nil
}

func (s *Store) ListReturns(ctx context.Context, storeID string, limit, offset int) ([]domain.SyncedReturn, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	returns := make([]domain.SyncedReturn, 0)
	for _, ret := range s.returns {
		if ret.StoreID == storeID {
			returns = append(returns, cloneSyncedReturn(ret))
		}
	}
	sort.Slice(returns, func(i, j int) bool {
		if returns[i].SettledAt.Equal(returns[j].SettledAt) {
			return returns[i].ID > returns[j].ID
		}
		return returns[i].SettledAt.After(returns[j].SettledAt)
	})

	total := len(returns)
	if offset >= total {
		return []domain.SyncedReturn{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return append([]domain.SyncedReturn(nil), returns[offset:end]...), total, nil
}

func (s *Store) SaveOperationalDay(ctx context.Context, day domain.SyncedOperationalDay) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.operationalDays[productKey(day.StoreID, day.ID)] = cloneSyncedOperationalDay(day)
	return nil
}

func (s *Store) FindOperationalDay(ctx context.Context, storeID string, operationalDayID string) (domain.SyncedOperationalDay, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	day, ok := s.operationalDays[productKey(storeID, operationalDayID)]
	if !ok {
		return domain.SyncedOperationalDay{}, app.ErrOperationalDayNotFound
	}
	return cloneSyncedOperationalDay(day), nil
}

func (s *Store) ListOperationalDays(ctx context.Context, storeID string, limit, offset int) ([]domain.SyncedOperationalDay, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	days := make([]domain.SyncedOperationalDay, 0)
	for _, day := range s.operationalDays {
		if day.StoreID == storeID {
			days = append(days, cloneSyncedOperationalDay(day))
		}
	}
	sort.Slice(days, func(i, j int) bool {
		if days[i].ClosedAt.Equal(days[j].ClosedAt) {
			return days[i].ID > days[j].ID
		}
		return days[i].ClosedAt.After(days[j].ClosedAt)
	})

	total := len(days)
	if offset >= total {
		return []domain.SyncedOperationalDay{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return append([]domain.SyncedOperationalDay(nil), days[offset:end]...), total, nil
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

func cloneSyncedPayment(payment domain.SyncedPayment) domain.SyncedPayment {
	return payment
}

func cloneSyncedCashMovement(movement domain.SyncedCashMovement) domain.SyncedCashMovement {
	return movement
}

func cloneSyncedFiscalDocument(document domain.SyncedFiscalDocument) domain.SyncedFiscalDocument {
	return document
}

func cloneSyncedReturn(ret domain.SyncedReturn) domain.SyncedReturn {
	cloned := ret
	if ret.PaymentIDs != nil {
		cloned.PaymentIDs = append([]string(nil), ret.PaymentIDs...)
	}
	return cloned
}

func cloneSyncedOperationalDay(day domain.SyncedOperationalDay) domain.SyncedOperationalDay {
	return day
}
