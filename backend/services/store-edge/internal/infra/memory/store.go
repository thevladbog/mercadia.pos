package memory

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
)

type Store struct {
	mu                 sync.RWMutex
	receipts           map[string]domain.Receipt
	terminals          map[string]domain.Terminal
	products           map[string]domain.Product
	barcodes           map[string]string
	payments           map[string]domain.Payment
	paymentsByReceipt  map[string][]string
	fiscalDocuments    map[string]domain.FiscalDocument
	fiscalByReceipt    map[string][]string
	fiscalByReturn     map[string]string
	cashMovements      map[string]domain.CashMovement
	cashByStore        map[string][]string
	cashRecounts       map[string]domain.CashRecount
	recountsByStore    map[string][]string
	shifts             map[string]domain.Shift
	shiftsByStore      map[string][]string
	operationalDays    map[string]domain.OperationalDay
	daysByStore        map[string][]string
	idempotency        map[string]app.IdempotencyRecord
	outboxEvents       map[string]domain.OutboxEvent
	actors             map[string]domain.Actor
	credentialPolicies map[string]domain.CredentialPolicy
	authSettings       map[string]domain.StoreAuthSettings
	authAttempts       []domain.AuthAttempt
	sessions           map[string]domain.Session
	returns            map[string]domain.Return
	returnsByStore     map[string][]string
	journalEntries     map[string]domain.OperationJournalEntry
	journalByStore     map[string][]string
	catalogSyncState   map[string]time.Time
}

type StoreOption func(*Store)

func NewStore(options ...StoreOption) *Store {
	store := &Store{
		receipts:           map[string]domain.Receipt{},
		terminals:          map[string]domain.Terminal{},
		products:           map[string]domain.Product{},
		barcodes:           map[string]string{},
		payments:           map[string]domain.Payment{},
		paymentsByReceipt:  map[string][]string{},
		fiscalDocuments:    map[string]domain.FiscalDocument{},
		fiscalByReceipt:    map[string][]string{},
		fiscalByReturn:     map[string]string{},
		cashMovements:      map[string]domain.CashMovement{},
		cashByStore:        map[string][]string{},
		cashRecounts:       map[string]domain.CashRecount{},
		recountsByStore:    map[string][]string{},
		shifts:             map[string]domain.Shift{},
		shiftsByStore:      map[string][]string{},
		operationalDays:    map[string]domain.OperationalDay{},
		daysByStore:        map[string][]string{},
		idempotency:        map[string]app.IdempotencyRecord{},
		outboxEvents:       map[string]domain.OutboxEvent{},
		actors:             map[string]domain.Actor{},
		credentialPolicies: map[string]domain.CredentialPolicy{},
		authSettings:       map[string]domain.StoreAuthSettings{},
		authAttempts:       []domain.AuthAttempt{},
		sessions:           map[string]domain.Session{},
		returns:            map[string]domain.Return{},
		returnsByStore:     map[string][]string{},
		journalEntries:     map[string]domain.OperationJournalEntry{},
		journalByStore:     map[string][]string{},
		catalogSyncState:   map[string]time.Time{},
	}
	for _, option := range options {
		option(store)
	}
	return store
}

func WithProducts(products ...domain.Product) StoreOption {
	return func(store *Store) {
		for _, product := range products {
			store.saveProduct(product)
		}
	}
}

func (s *Store) SaveReceipt(ctx context.Context, receipt domain.Receipt) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.receipts[receipt.ID] = cloneReceipt(receipt)
	return nil
}

func (s *Store) FindReceipt(ctx context.Context, receiptID string) (domain.Receipt, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	receipt, ok := s.receipts[receiptID]
	if !ok {
		return domain.Receipt{}, app.ErrReceiptNotFound
	}
	return cloneReceipt(receipt), nil
}

func (s *Store) ListReceiptsByShift(ctx context.Context, shiftID string) ([]domain.Receipt, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	receipts := []domain.Receipt{}
	for _, receipt := range s.receipts {
		if receipt.ShiftID == shiftID {
			receipts = append(receipts, cloneReceipt(receipt))
		}
	}
	return receipts, nil
}

func (s *Store) ListReceiptsByOperationalDay(ctx context.Context, operationalDayID string) ([]domain.Receipt, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	receipts := []domain.Receipt{}
	for _, receipt := range s.receipts {
		if receipt.OperationalDayID == operationalDayID {
			receipts = append(receipts, cloneReceipt(receipt))
		}
	}
	return receipts, nil
}

func (s *Store) SaveTerminal(ctx context.Context, terminal domain.Terminal) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.terminals[terminal.ID] = terminal
	return nil
}

func (s *Store) FindTerminal(ctx context.Context, terminalID string) (domain.Terminal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	terminal, ok := s.terminals[terminalID]
	if !ok {
		return domain.Terminal{}, app.ErrTerminalNotFound
	}
	return terminal, nil
}

func (s *Store) ListTerminalsByStore(ctx context.Context, storeID string) ([]domain.Terminal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	terminals := make([]domain.Terminal, 0)
	for _, terminal := range s.terminals {
		if terminal.StoreID == storeID {
			terminals = append(terminals, terminal)
		}
	}
	sort.Slice(terminals, func(i, j int) bool {
		return terminals[i].ID < terminals[j].ID
	})
	return terminals, nil
}

func (s *Store) FindProductByBarcode(ctx context.Context, barcode string) (domain.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	productID, ok := s.barcodes[barcode]
	if !ok {
		return domain.Product{}, app.ErrProductNotFound
	}
	product, ok := s.products[productID]
	if !ok || !product.Active {
		return domain.Product{}, app.ErrProductNotFound
	}
	return cloneProduct(product), nil
}

func (s *Store) SaveProduct(_ context.Context, product domain.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.saveProduct(product)
	return nil
}

func (s *Store) GetLastSyncedAt(_ context.Context, storeID string) (time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.catalogSyncState[storeID], nil
}

func (s *Store) SaveLastSyncedAt(_ context.Context, storeID string, syncedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.catalogSyncState[storeID] = syncedAt.UTC()
	return nil
}

func (s *Store) SavePayment(ctx context.Context, payment domain.Payment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.payments[payment.ID]; !exists {
		s.paymentsByReceipt[payment.ReceiptID] = append(s.paymentsByReceipt[payment.ReceiptID], payment.ID)
	}
	s.payments[payment.ID] = payment
	return nil
}

func (s *Store) FindPayment(ctx context.Context, paymentID string) (domain.Payment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	payment, ok := s.payments[paymentID]
	if !ok {
		return domain.Payment{}, app.ErrPaymentNotFound
	}
	return payment, nil
}

func (s *Store) FindPaymentsByReceipt(ctx context.Context, receiptID string) ([]domain.Payment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	paymentIDs := s.paymentsByReceipt[receiptID]
	payments := make([]domain.Payment, 0, len(paymentIDs))
	for _, paymentID := range paymentIDs {
		payment, ok := s.payments[paymentID]
		if ok {
			payments = append(payments, payment)
		}
	}
	return payments, nil
}

func (s *Store) CountFiscalizedReceiptsByStoreAndBusinessDate(ctx context.Context, storeID string, businessDate string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int
	for _, receipt := range s.receipts {
		receiptBusinessDate := receipt.BusinessDate
		if receiptBusinessDate == "" {
			receiptBusinessDate = receipt.CreatedAt.UTC().Format("2006-01-02")
		}
		if receipt.StoreID == storeID &&
			receipt.Status == domain.ReceiptStatusFiscalized &&
			receiptBusinessDate == businessDate {
			count++
		}
	}
	return count, nil
}

func (s *Store) ListUnresolvedReceiptsByStoreAndBusinessDate(ctx context.Context, storeID string, businessDate string) ([]domain.Receipt, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	receipts := []domain.Receipt{}
	for _, receipt := range s.receipts {
		receiptBusinessDate := receipt.BusinessDate
		if receiptBusinessDate == "" {
			receiptBusinessDate = receipt.CreatedAt.UTC().Format("2006-01-02")
		}
		if receipt.StoreID != storeID || receiptBusinessDate != businessDate {
			continue
		}
		if receipt.Status == domain.ReceiptStatusDraft ||
			receipt.Status == domain.ReceiptStatusPaymentStarted ||
			receipt.Status == domain.ReceiptStatusPaid {
			receipts = append(receipts, cloneReceipt(receipt))
		}
	}
	return receipts, nil
}

func (s *Store) ListUnresolvedReceiptsByShift(ctx context.Context, shiftID string) ([]domain.Receipt, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	receipts := []domain.Receipt{}
	for _, receipt := range s.receipts {
		if receipt.ShiftID != shiftID {
			continue
		}
		if receipt.Status == domain.ReceiptStatusDraft ||
			receipt.Status == domain.ReceiptStatusPaymentStarted ||
			receipt.Status == domain.ReceiptStatusPaid {
			receipts = append(receipts, cloneReceipt(receipt))
		}
	}
	return receipts, nil
}

func (s *Store) SaveFiscalDocument(ctx context.Context, document domain.FiscalDocument) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.fiscalDocuments[document.ID]; !exists {
		s.fiscalByReceipt[document.ReceiptID] = append(s.fiscalByReceipt[document.ReceiptID], document.ID)
		if document.ReturnID != "" {
			s.fiscalByReturn[document.ReturnID] = document.ID
		}
	}
	s.fiscalDocuments[document.ID] = document
	return nil
}

func (s *Store) FindFiscalDocumentByReturn(ctx context.Context, returnID string) (domain.FiscalDocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	documentID, ok := s.fiscalByReturn[returnID]
	if !ok {
		return domain.FiscalDocument{}, app.ErrFiscalDocumentNotFound
	}
	document, ok := s.fiscalDocuments[documentID]
	if !ok {
		return domain.FiscalDocument{}, app.ErrFiscalDocumentNotFound
	}
	return document, nil
}

func (s *Store) FindFiscalDocumentsByReceipt(ctx context.Context, receiptID string) ([]domain.FiscalDocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	documentIDs := s.fiscalByReceipt[receiptID]
	documents := make([]domain.FiscalDocument, 0, len(documentIDs))
	for _, documentID := range documentIDs {
		document, ok := s.fiscalDocuments[documentID]
		if ok {
			documents = append(documents, document)
		}
	}
	return documents, nil
}

func (s *Store) SaveCashMovement(ctx context.Context, movement domain.CashMovement) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.cashMovements[movement.ID]; !exists {
		s.cashByStore[movement.StoreID] = append(s.cashByStore[movement.StoreID], movement.ID)
	}
	s.cashMovements[movement.ID] = movement
	return nil
}

func (s *Store) ListCashMovements(ctx context.Context, storeID string) ([]domain.CashMovement, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	movementIDs := s.cashByStore[storeID]
	movements := make([]domain.CashMovement, 0, len(movementIDs))
	for _, movementID := range movementIDs {
		movement, ok := s.cashMovements[movementID]
		if ok {
			movements = append(movements, movement)
		}
	}
	return movements, nil
}

func (s *Store) SaveCashRecount(ctx context.Context, recount domain.CashRecount) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.cashRecounts[recount.ID]; !exists {
		s.recountsByStore[recount.StoreID] = append(s.recountsByStore[recount.StoreID], recount.ID)
	}
	s.cashRecounts[recount.ID] = recount
	return nil
}

func (s *Store) FindCashRecount(ctx context.Context, recountID string) (domain.CashRecount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	recount, ok := s.cashRecounts[recountID]
	if !ok {
		return domain.CashRecount{}, app.ErrCashRecountNotFound
	}
	return recount, nil
}

func (s *Store) ListCashRecounts(ctx context.Context, storeID string) ([]domain.CashRecount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	recountIDs := s.recountsByStore[storeID]
	recounts := make([]domain.CashRecount, 0, len(recountIDs))
	for _, recountID := range recountIDs {
		recount, ok := s.cashRecounts[recountID]
		if ok {
			recounts = append(recounts, recount)
		}
	}
	return recounts, nil
}

func (s *Store) ListCashRecountsByStoreAndBusinessDate(ctx context.Context, storeID string, businessDate string) ([]domain.CashRecount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	recountIDs := s.recountsByStore[storeID]
	recounts := make([]domain.CashRecount, 0, len(recountIDs))
	for _, recountID := range recountIDs {
		recount, ok := s.cashRecounts[recountID]
		if !ok {
			continue
		}
		recountBusinessDate := recount.BusinessDate
		if recountBusinessDate == "" {
			recountBusinessDate = recount.CreatedAt.UTC().Format("2006-01-02")
		}
		if recountBusinessDate == businessDate {
			recounts = append(recounts, recount)
		}
	}
	return recounts, nil
}

func (s *Store) ListUnresolvedCashRecountDiscrepanciesByStoreAndBusinessDate(ctx context.Context, storeID string, businessDate string) ([]domain.CashRecount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	recountIDs := s.recountsByStore[storeID]
	recounts := make([]domain.CashRecount, 0, len(recountIDs))
	for _, recountID := range recountIDs {
		recount, ok := s.cashRecounts[recountID]
		if !ok {
			continue
		}
		recountBusinessDate := recount.BusinessDate
		if recountBusinessDate == "" {
			recountBusinessDate = recount.CreatedAt.UTC().Format("2006-01-02")
		}
		if ok &&
			recount.Status == domain.CashRecountStatusDiscrepancy &&
			recount.ResolutionStatus == domain.CashRecountResolutionStatusOpen &&
			recountBusinessDate == businessDate {
			recounts = append(recounts, recount)
		}
	}
	return recounts, nil
}

func (s *Store) SaveShift(ctx context.Context, shift domain.Shift) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.shifts[shift.ID]; !exists {
		s.shiftsByStore[shift.StoreID] = append(s.shiftsByStore[shift.StoreID], shift.ID)
	}
	s.shifts[shift.ID] = shift
	return nil
}

func (s *Store) FindShift(ctx context.Context, shiftID string) (domain.Shift, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	shift, ok := s.shifts[shiftID]
	if !ok {
		return domain.Shift{}, app.ErrShiftNotFound
	}
	return shift, nil
}

func (s *Store) FindOpenShiftByTerminal(ctx context.Context, terminalID string) (domain.Shift, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, shift := range s.shifts {
		if shift.TerminalID == terminalID && shift.Status == domain.ShiftStatusOpen {
			return shift, nil
		}
	}
	return domain.Shift{}, app.ErrShiftNotFound
}

func (s *Store) FindOpenShiftByCashier(ctx context.Context, cashierID string) (domain.Shift, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, shift := range s.shifts {
		if shift.CashierID == cashierID && shift.Status == domain.ShiftStatusOpen {
			return shift, nil
		}
	}
	return domain.Shift{}, app.ErrShiftNotFound
}

func (s *Store) ListOpenShiftsByStore(ctx context.Context, storeID string) ([]domain.Shift, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	shiftIDs := s.shiftsByStore[storeID]
	shifts := make([]domain.Shift, 0, len(shiftIDs))
	for _, shiftID := range shiftIDs {
		shift, ok := s.shifts[shiftID]
		if ok && shift.Status == domain.ShiftStatusOpen {
			shifts = append(shifts, shift)
		}
	}
	return shifts, nil
}

func (s *Store) ListShiftsByOperationalDay(ctx context.Context, operationalDayID string) ([]domain.Shift, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	shifts := []domain.Shift{}
	for _, shift := range s.shifts {
		if shift.OperationalDayID == operationalDayID {
			shifts = append(shifts, shift)
		}
	}
	return shifts, nil
}

func (s *Store) SaveOperationalDay(ctx context.Context, day domain.OperationalDay) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.operationalDays[day.ID]; !exists {
		s.daysByStore[day.StoreID] = append(s.daysByStore[day.StoreID], day.ID)
	}
	s.operationalDays[day.ID] = day
	return nil
}

func (s *Store) FindOperationalDay(ctx context.Context, dayID string) (domain.OperationalDay, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	day, ok := s.operationalDays[dayID]
	if !ok {
		return domain.OperationalDay{}, app.ErrOperationalDayNotFound
	}
	return day, nil
}

func (s *Store) FindOpenOperationalDayByStore(ctx context.Context, storeID string) (domain.OperationalDay, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, dayID := range s.daysByStore[storeID] {
		day, ok := s.operationalDays[dayID]
		if ok && day.Status == domain.OperationalDayStatusOpen {
			return day, nil
		}
	}
	return domain.OperationalDay{}, app.ErrOperationalDayNotFound
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

func idempotencyMapKey(operation string, key string) string {
	return operation + "\x00" + key
}

func cloneReceipt(receipt domain.Receipt) domain.Receipt {
	receipt.Lines = append([]domain.ReceiptLine(nil), receipt.Lines...)
	return receipt
}

func (s *Store) saveProduct(product domain.Product) {
	product = cloneProduct(product)
	s.products[product.ID] = product
	for _, barcode := range product.Barcodes {
		s.barcodes[barcode] = product.ID
	}
}

func cloneProduct(product domain.Product) domain.Product {
	product.Barcodes = append([]string(nil), product.Barcodes...)
	return product
}

func (s *Store) SaveOutboxEvent(ctx context.Context, event domain.OutboxEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.outboxEvents[event.ID] = cloneOutboxEvent(event)
	return nil
}

func (s *Store) ListPendingOutboxEvents(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	events := make([]domain.OutboxEvent, 0, limit)
	for _, event := range s.outboxEvents {
		if event.PublishedAt != nil {
			continue
		}
		events = append(events, cloneOutboxEvent(event))
		if len(events) >= limit {
			break
		}
	}
	return events, nil
}

func (s *Store) MarkOutboxEventPublished(ctx context.Context, eventID string, publishedAt time.Time) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.outboxEvents[eventID]
	if !ok {
		return false, nil
	}
	if event.PublishedAt != nil {
		return false, nil
	}
	publishedAt = publishedAt.UTC()
	event.PublishedAt = &publishedAt
	s.outboxEvents[eventID] = event
	return true, nil
}

func (s *Store) CountOutboxEvents(ctx context.Context) (pending int64, published int64, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, event := range s.outboxEvents {
		if event.PublishedAt == nil {
			pending++
		} else {
			published++
		}
	}
	return pending, published, nil
}

func cloneOutboxEvent(event domain.OutboxEvent) domain.OutboxEvent {
	if event.Payload != nil {
		event.Payload = append(json.RawMessage(nil), event.Payload...)
	}
	if event.PublishedAt != nil {
		publishedAt := *event.PublishedAt
		event.PublishedAt = &publishedAt
	}
	return event
}

func (s *Store) Run(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}
