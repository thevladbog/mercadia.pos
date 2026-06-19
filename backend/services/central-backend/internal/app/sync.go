package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/domain"
)

var (
	ErrInvalidSyncCommand = errors.New("invalid sync command")
	ErrSyncEventDuplicate = errors.New("sync event already accepted")
)

type SyncEventRepository interface {
	SaveSyncEvent(ctx context.Context, event domain.SyncEvent) error
	ExistsSyncEvent(ctx context.Context, storeID string, sourceEventID string) (bool, error)
	ListSyncEvents(ctx context.Context, storeID string, limit, offset int) ([]domain.SyncEvent, int, error)
}

type SyncEventInput struct {
	EventID    string
	EventType  string
	OccurredAt time.Time
	Payload    json.RawMessage
}

type AcceptSyncEventsCommand struct {
	StoreID        string
	IdempotencyKey string
	Events         []SyncEventInput
}

type SyncEventsResult struct {
	StoreID  string
	Status   string
	Accepted int
}

type SyncService struct {
	stores          StoreRepository
	syncEvents      SyncEventRepository
	catalog         CatalogProductRepository
	payments        SyncedPaymentRepository
	cashMovements   SyncedCashMovementRepository
	fiscalDocs      SyncedFiscalDocumentRepository
	returns         SyncedReturnRepository
	operationalDays SyncedOperationalDayRepository
	idempotency     IdempotencyStore
	now             func() time.Time
	newID           func(prefix string) string
}

func NewSyncService(
	stores StoreRepository,
	syncEvents SyncEventRepository,
	catalog CatalogProductRepository,
	payments SyncedPaymentRepository,
	cashMovements SyncedCashMovementRepository,
	fiscalDocs SyncedFiscalDocumentRepository,
	returns SyncedReturnRepository,
	operationalDays SyncedOperationalDayRepository,
	idempotency IdempotencyStore,
) *SyncService {
	return &SyncService{
		stores:          stores,
		syncEvents:      syncEvents,
		catalog:         catalog,
		payments:        payments,
		cashMovements:   cashMovements,
		fiscalDocs:      fiscalDocs,
		returns:         returns,
		operationalDays: operationalDays,
		idempotency:     idempotency,
		now:             time.Now,
		newID:           defaultNewID,
	}
}

func (s *SyncService) AcceptEvents(ctx context.Context, command AcceptSyncEventsCommand) (SyncEventsResult, error) {
	const operation = "accept_sync_events"
	if command.StoreID == "" || len(command.Events) == 0 {
		return SyncEventsResult{}, ErrInvalidSyncCommand
	}
	if command.IdempotencyKey == "" {
		return SyncEventsResult{}, ErrIdempotencyKeyRequired
	}

	if _, err := s.stores.FindStore(ctx, command.StoreID); err != nil {
		return SyncEventsResult{}, err
	}

	fingerprint := syncEventsFingerprint(command)
	if result, found, err := s.findSyncIdempotency(ctx, operation, command.IdempotencyKey, command.StoreID, fingerprint); err != nil || found {
		return result, err
	}

	accepted := 0
	now := s.now().UTC()
	for _, input := range command.Events {
		if input.EventID == "" || input.EventType == "" {
			return SyncEventsResult{}, ErrInvalidSyncCommand
		}
		exists, err := s.syncEvents.ExistsSyncEvent(ctx, command.StoreID, input.EventID)
		if err != nil {
			return SyncEventsResult{}, err
		}
		if exists {
			continue
		}

		occurredAt := input.OccurredAt
		if occurredAt.IsZero() {
			occurredAt = now
		}

		event, err := domain.NewSyncEvent(domain.SyncEvent{
			ID:            s.newID("sync"),
			StoreID:       command.StoreID,
			EventType:     input.EventType,
			SourceEventID: input.EventID,
			Payload:       input.Payload,
			OccurredAt:    occurredAt.UTC(),
			ReceivedAt:    now,
		})
		if err != nil {
			return SyncEventsResult{}, ErrInvalidSyncCommand
		}
		if err := s.syncEvents.SaveSyncEvent(ctx, event); err != nil {
			if errors.Is(err, ErrSyncEventDuplicate) {
				continue
			}
			return SyncEventsResult{}, err
		}
		if err := s.applySyncEvent(ctx, event); err != nil {
			return SyncEventsResult{}, err
		}
		accepted++
	}

	result := SyncEventsResult{
		StoreID:  command.StoreID,
		Status:   "accepted",
		Accepted: accepted,
	}
	if err := s.saveSyncIdempotency(ctx, operation, command.IdempotencyKey, command.StoreID, fingerprint, result); err != nil {
		return SyncEventsResult{}, err
	}
	return result, nil
}

func (s *SyncService) ListEvents(ctx context.Context, storeID string, params PageParams) (PageResult[domain.SyncEvent], error) {
	if storeID == "" {
		return PageResult[domain.SyncEvent]{}, ErrInvalidSyncCommand
	}
	if _, err := s.stores.FindStore(ctx, storeID); err != nil {
		return PageResult[domain.SyncEvent]{}, err
	}

	events, total, err := s.syncEvents.ListSyncEvents(ctx, storeID, params.Limit, params.Offset)
	if err != nil {
		return PageResult[domain.SyncEvent]{}, err
	}
	return PageResult[domain.SyncEvent]{Items: events, TotalCount: total}, nil
}

func (s *SyncService) applySyncEvent(ctx context.Context, event domain.SyncEvent) error {
	switch event.EventType {
	case "catalog.product.upserted":
		return s.upsertCatalogProductFromPayload(ctx, event)
	case "payment.captured":
		return s.upsertPaymentFromPayload(ctx, event)
	case "payment.cancelled":
		return s.updatePaymentCancelledFromPayload(ctx, event)
	case "payment.refunded":
		return s.updatePaymentRefundedFromPayload(ctx, event)
	case "cash.movement.posted":
		return s.upsertCashMovementFromPayload(ctx, event)
	case "fiscal.document.created":
		return s.upsertFiscalDocumentFromPayload(ctx, event)
	case "return.settled":
		return s.upsertReturnFromPayload(ctx, event)
	case "operational_day.closed":
		return s.upsertOperationalDayFromPayload(ctx, event)
	default:
		return nil
	}
}

func (s *SyncService) upsertCatalogProductFromPayload(ctx context.Context, event domain.SyncEvent) error {
	var payload struct {
		ProductID      string   `json:"productId"`
		Name           string   `json:"name"`
		Barcodes       []string `json:"barcodes"`
		UnitPriceMinor int64    `json:"unitPriceMinor"`
		TaxCategoryID  string   `json:"taxCategoryId"`
		Active         *bool    `json:"active"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return ErrInvalidSyncCommand
	}

	active := true
	if payload.Active != nil {
		active = *payload.Active
	}

	existing, err := s.catalog.FindProduct(ctx, event.StoreID, payload.ProductID)
	version := int64(1)
	if err == nil {
		version = existing.Version + 1
	} else if !errors.Is(err, ErrCatalogProductNotFound) {
		return err
	}

	product, err := domain.NewCatalogProduct(domain.CatalogProduct{
		ID:             payload.ProductID,
		StoreID:        event.StoreID,
		Name:           payload.Name,
		Barcodes:       payload.Barcodes,
		UnitPriceMinor: payload.UnitPriceMinor,
		TaxCategoryID:  payload.TaxCategoryID,
		Active:         active,
		Version:        version,
		UpdatedAt:      event.ReceivedAt,
	})
	if err != nil {
		return ErrInvalidSyncCommand
	}
	return s.catalog.SaveProduct(ctx, product)
}

func (s *SyncService) upsertPaymentFromPayload(ctx context.Context, event domain.SyncEvent) error {
	var payload struct {
		PaymentID   string    `json:"paymentId"`
		ReceiptID   string    `json:"receiptId"`
		Method      string    `json:"method"`
		AmountMinor int64     `json:"amountMinor"`
		CapturedAt  time.Time `json:"capturedAt"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return ErrInvalidSyncCommand
	}

	payment, err := domain.NewSyncedPayment(domain.SyncedPayment{
		ID:            payload.PaymentID,
		StoreID:       event.StoreID,
		ReceiptID:     payload.ReceiptID,
		Method:        payload.Method,
		AmountMinor:   payload.AmountMinor,
		Status:        domain.SyncedPaymentStatusCaptured,
		CapturedAt:    payload.CapturedAt.UTC(),
		SourceEventID: event.SourceEventID,
		LastEventID:   event.SourceEventID,
		SyncedAt:      event.ReceivedAt,
		UpdatedAt:     event.ReceivedAt,
	})
	if err != nil {
		return ErrInvalidSyncCommand
	}
	return s.payments.SavePayment(ctx, payment)
}

func (s *SyncService) updatePaymentCancelledFromPayload(ctx context.Context, event domain.SyncEvent) error {
	var payload struct {
		PaymentID   string    `json:"paymentId"`
		ReceiptID   string    `json:"receiptId"`
		Method      string    `json:"method"`
		AmountMinor int64     `json:"amountMinor"`
		CancelledAt time.Time `json:"cancelledAt"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return ErrInvalidSyncCommand
	}

	cancelledAt := payload.CancelledAt.UTC()
	now := event.ReceivedAt

	existing, err := s.payments.FindPayment(ctx, event.StoreID, payload.PaymentID)
	if err != nil {
		if !errors.Is(err, ErrPaymentNotFound) {
			return err
		}
		payment, err := domain.NewSyncedPayment(domain.SyncedPayment{
			ID:            payload.PaymentID,
			StoreID:       event.StoreID,
			ReceiptID:     payload.ReceiptID,
			Method:        payload.Method,
			AmountMinor:   payload.AmountMinor,
			Status:        domain.SyncedPaymentStatusCancelled,
			CapturedAt:    cancelledAt,
			CancelledAt:   &cancelledAt,
			SourceEventID: event.SourceEventID,
			LastEventID:   event.SourceEventID,
			SyncedAt:      now,
			UpdatedAt:     now,
		})
		if err != nil {
			return ErrInvalidSyncCommand
		}
		return s.payments.SavePayment(ctx, payment)
	}

	existing.Status = domain.SyncedPaymentStatusCancelled
	existing.CancelledAt = &cancelledAt
	existing.LastEventID = event.SourceEventID
	existing.UpdatedAt = now
	_, err = domain.NewSyncedPayment(existing)
	if err != nil {
		return ErrInvalidSyncCommand
	}
	return s.payments.SavePayment(ctx, existing)
}

func (s *SyncService) updatePaymentRefundedFromPayload(ctx context.Context, event domain.SyncEvent) error {
	var payload struct {
		PaymentID            string    `json:"paymentId"`
		ReceiptID            string    `json:"receiptId"`
		Method               string    `json:"method"`
		AmountMinor          int64     `json:"amountMinor"`
		RefundedAmountMinor  int64     `json:"refundedAmountMinor"`
		RemainingAmountMinor int64     `json:"remainingAmountMinor"`
		RefundedAt           time.Time `json:"refundedAt"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return ErrInvalidSyncCommand
	}

	status := domain.SyncedPaymentStatusPartiallyRefunded
	if payload.RemainingAmountMinor == 0 {
		status = domain.SyncedPaymentStatusRefunded
	}
	refundedAt := payload.RefundedAt.UTC()
	now := event.ReceivedAt

	existing, err := s.payments.FindPayment(ctx, event.StoreID, payload.PaymentID)
	if err != nil {
		if !errors.Is(err, ErrPaymentNotFound) {
			return err
		}
		payment, err := domain.NewSyncedPayment(domain.SyncedPayment{
			ID:                   payload.PaymentID,
			StoreID:              event.StoreID,
			ReceiptID:            payload.ReceiptID,
			Method:               payload.Method,
			AmountMinor:          payload.AmountMinor,
			Status:               status,
			CapturedAt:           refundedAt,
			RefundedAmountMinor:  payload.RefundedAmountMinor,
			RemainingAmountMinor: payload.RemainingAmountMinor,
			SourceEventID:        event.SourceEventID,
			LastEventID:          event.SourceEventID,
			SyncedAt:             now,
			UpdatedAt:            refundedAt,
		})
		if err != nil {
			return ErrInvalidSyncCommand
		}
		return s.payments.SavePayment(ctx, payment)
	}

	existing.Status = status
	existing.RefundedAmountMinor = payload.RefundedAmountMinor
	existing.RemainingAmountMinor = payload.RemainingAmountMinor
	existing.LastEventID = event.SourceEventID
	existing.UpdatedAt = refundedAt
	_, err = domain.NewSyncedPayment(existing)
	if err != nil {
		return ErrInvalidSyncCommand
	}
	return s.payments.SavePayment(ctx, existing)
}

func (s *SyncService) upsertCashMovementFromPayload(ctx context.Context, event domain.SyncEvent) error {
	var payload struct {
		CashMovementID    string    `json:"cashMovementId"`
		Type              string    `json:"type"`
		FromContainerID   string    `json:"fromContainerId"`
		FromContainerType string    `json:"fromContainerType"`
		ToContainerID     string    `json:"toContainerId"`
		ToContainerType   string    `json:"toContainerType"`
		AmountMinor       int64     `json:"amountMinor"`
		Currency          string    `json:"currency"`
		ActorID           string    `json:"actorId"`
		PostedAt          time.Time `json:"postedAt"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return ErrInvalidSyncCommand
	}

	movement, err := domain.NewSyncedCashMovement(domain.SyncedCashMovement{
		ID:                payload.CashMovementID,
		StoreID:           event.StoreID,
		Type:              payload.Type,
		FromContainerID:   payload.FromContainerID,
		FromContainerType: payload.FromContainerType,
		ToContainerID:     payload.ToContainerID,
		ToContainerType:   payload.ToContainerType,
		AmountMinor:       payload.AmountMinor,
		Currency:          payload.Currency,
		ActorID:           payload.ActorID,
		PostedAt:          payload.PostedAt.UTC(),
		SourceEventID:     event.SourceEventID,
		SyncedAt:          event.ReceivedAt,
	})
	if err != nil {
		return ErrInvalidSyncCommand
	}
	return s.cashMovements.SaveCashMovement(ctx, movement)
}

func (s *SyncService) upsertFiscalDocumentFromPayload(ctx context.Context, event domain.SyncEvent) error {
	var payload struct {
		FiscalDocumentID string    `json:"fiscalDocumentId"`
		ReceiptID        string    `json:"receiptId"`
		Kind             string    `json:"kind"`
		AmountMinor      int64     `json:"amountMinor"`
		DeviceID         string    `json:"deviceId"`
		FiscalSign       string    `json:"fiscalSign"`
		FiscalizedAt     time.Time `json:"fiscalizedAt"`
		ReturnID         string    `json:"returnId"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return ErrInvalidSyncCommand
	}

	document, err := domain.NewSyncedFiscalDocument(domain.SyncedFiscalDocument{
		ID:            payload.FiscalDocumentID,
		StoreID:       event.StoreID,
		ReceiptID:     payload.ReceiptID,
		Kind:          payload.Kind,
		AmountMinor:   payload.AmountMinor,
		DeviceID:      payload.DeviceID,
		FiscalSign:    payload.FiscalSign,
		FiscalizedAt:  payload.FiscalizedAt.UTC(),
		ReturnID:      payload.ReturnID,
		SourceEventID: event.SourceEventID,
		SyncedAt:      event.ReceivedAt,
	})
	if err != nil {
		return ErrInvalidSyncCommand
	}
	return s.fiscalDocs.SaveFiscalDocument(ctx, document)
}

func (s *SyncService) upsertReturnFromPayload(ctx context.Context, event domain.SyncEvent) error {
	var payload struct {
		ReturnID       string    `json:"returnId"`
		ReceiptID      string    `json:"receiptId"`
		TotalMinor     int64     `json:"totalMinor"`
		PaymentIDs     []string  `json:"paymentIds"`
		CashMovementID string    `json:"cashMovementId"`
		SettledAt      time.Time `json:"settledAt"`
		ActorID        string    `json:"actorId"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return ErrInvalidSyncCommand
	}

	ret, err := domain.NewSyncedReturn(domain.SyncedReturn{
		ID:             payload.ReturnID,
		StoreID:        event.StoreID,
		ReceiptID:      payload.ReceiptID,
		TotalMinor:     payload.TotalMinor,
		PaymentIDs:     payload.PaymentIDs,
		CashMovementID: payload.CashMovementID,
		ActorID:        payload.ActorID,
		SettledAt:      payload.SettledAt.UTC(),
		SourceEventID:  event.SourceEventID,
		SyncedAt:       event.ReceivedAt,
	})
	if err != nil {
		return ErrInvalidSyncCommand
	}
	return s.returns.SaveReturn(ctx, ret)
}

func (s *SyncService) upsertOperationalDayFromPayload(ctx context.Context, event domain.SyncEvent) error {
	var payload struct {
		OperationalDayID string    `json:"operationalDayId"`
		BusinessDate     string    `json:"businessDate"`
		ClosedByID       string    `json:"closedById"`
		ClosedAt         time.Time `json:"closedAt"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return ErrInvalidSyncCommand
	}

	day, err := domain.NewSyncedOperationalDay(domain.SyncedOperationalDay{
		ID:            payload.OperationalDayID,
		StoreID:       event.StoreID,
		BusinessDate:  payload.BusinessDate,
		ClosedByID:    payload.ClosedByID,
		ClosedAt:      payload.ClosedAt.UTC(),
		SourceEventID: event.SourceEventID,
		SyncedAt:      event.ReceivedAt,
	})
	if err != nil {
		return ErrInvalidSyncCommand
	}
	return s.operationalDays.SaveOperationalDay(ctx, day)
}

func (s *SyncService) findSyncIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string) (SyncEventsResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, operation, key)
	if err != nil || !found {
		return SyncEventsResult{}, found, err
	}
	if record.Fingerprint != fingerprint {
		return SyncEventsResult{}, true, ErrIdempotencyKeyReused
	}
	if record.TargetID != "" && targetID != "" && record.TargetID != targetID {
		return SyncEventsResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(SyncEventsResult)
	if !ok {
		return SyncEventsResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

func (s *SyncService) saveSyncIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string, result SyncEventsResult) error {
	return s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         key,
		TargetID:    targetID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now().UTC(),
	})
}

func syncEventsFingerprint(command AcceptSyncEventsCommand) string {
	sum := sha256.New()
	_, _ = fmt.Fprintf(sum, "%s|", command.StoreID)
	for _, event := range command.Events {
		_, _ = fmt.Fprintf(sum, "%s|%s|%s|", event.EventID, event.EventType, string(event.Payload))
	}
	return hex.EncodeToString(sum.Sum(nil))
}
