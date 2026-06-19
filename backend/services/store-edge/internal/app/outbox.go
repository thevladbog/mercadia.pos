package app

import (
	"context"
	"encoding/json"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

type OutboxRepository interface {
	SaveOutboxEvent(ctx context.Context, event domain.OutboxEvent) error
	ListPendingOutboxEvents(ctx context.Context, limit int) ([]domain.OutboxEvent, error)
	MarkOutboxEventPublished(ctx context.Context, eventID string, publishedAt time.Time) (bool, error)
	CountOutboxEvents(ctx context.Context) (pending int64, published int64, err error)
}

type OutboxRecorder interface {
	RecordPaymentCaptured(ctx context.Context, payment domain.Payment, storeID string) error
	RecordPaymentCancelled(ctx context.Context, payment domain.Payment, storeID string, actorID string, reason string) error
	RecordPaymentRefunded(ctx context.Context, payment domain.Payment, storeID string, actorID string, reason string) error
	RecordFiscalDocumentCreated(ctx context.Context, document domain.FiscalDocument, storeID string) error
	RecordCashMovementPosted(ctx context.Context, movement domain.CashMovement) error
	RecordOperationalDayClosed(ctx context.Context, day domain.OperationalDay) error
	RecordReturnSettled(ctx context.Context, ret domain.Return, paymentIDs []string, storeID string, actorID string) error
}

type OutboxService struct {
	repo  OutboxRepository
	now   func() time.Time
	newID func(prefix string) string
}

type OutboxOption func(*OutboxService)

func NewOutboxService(repo OutboxRepository, options ...OutboxOption) *OutboxService {
	service := &OutboxService{
		repo: repo,
		now: func() time.Time {
			return time.Now().UTC()
		},
		newID: randomID,
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithOutboxClock(now func() time.Time) OutboxOption {
	return func(service *OutboxService) {
		service.now = now
	}
}

func WithOutboxIDGenerator(newID func(prefix string) string) OutboxOption {
	return func(service *OutboxService) {
		service.newID = newID
	}
}

type OutboxStatus struct {
	PendingCount    int64
	PublishedCount  int64
	BrokerConnected bool
}

func (s *OutboxService) Enqueue(ctx context.Context, event domain.OutboxEvent) error {
	event, err := domain.NewOutboxEvent(event)
	if err != nil {
		return err
	}
	return s.repo.SaveOutboxEvent(ctx, event)
}

func (s *OutboxService) Status(ctx context.Context, brokerConnected bool) (OutboxStatus, error) {
	pending, published, err := s.repo.CountOutboxEvents(ctx)
	if err != nil {
		return OutboxStatus{}, err
	}
	return OutboxStatus{
		PendingCount:    pending,
		PublishedCount:  published,
		BrokerConnected: brokerConnected,
	}, nil
}

func (s *OutboxService) RecordPaymentCaptured(ctx context.Context, payment domain.Payment, storeID string) error {
	payload, err := json.Marshal(map[string]any{
		"storeId":   storeID,
		"paymentId": payment.ID,
		"receiptId": payment.ReceiptID,
		"method":    payment.Method,
		"amountMinor": payment.AmountMinor,
		"capturedAt": payment.CapturedAt,
	})
	if err != nil {
		return err
	}
	return s.Enqueue(ctx, domain.OutboxEvent{
		ID:            s.newID("obx"),
		AggregateType: domain.OutboxAggregatePayment,
		AggregateID:   payment.ID,
		EventType:     domain.OutboxEventPaymentCaptured,
		Payload:       payload,
		CreatedAt:     s.now(),
	})
}

func (s *OutboxService) RecordPaymentCancelled(ctx context.Context, payment domain.Payment, storeID string, actorID string, reason string) error {
	payload, err := json.Marshal(map[string]any{
		"storeId":   storeID,
		"paymentId": payment.ID,
		"receiptId": payment.ReceiptID,
		"method":    payment.Method,
		"amountMinor": payment.AmountMinor,
		"cancelledAt": payment.UpdatedAt,
		"actorId":   actorID,
		"reason":    reason,
	})
	if err != nil {
		return err
	}
	return s.Enqueue(ctx, domain.OutboxEvent{
		ID:            s.newID("obx"),
		AggregateType: domain.OutboxAggregatePayment,
		AggregateID:   payment.ID,
		EventType:     domain.OutboxEventPaymentCancelled,
		Payload:       payload,
		CreatedAt:     s.now(),
	})
}

func (s *OutboxService) RecordPaymentRefunded(ctx context.Context, payment domain.Payment, storeID string, actorID string, reason string) error {
	payload, err := json.Marshal(map[string]any{
		"storeId":     storeID,
		"paymentId":   payment.ID,
		"receiptId":   payment.ReceiptID,
		"method":      payment.Method,
		"amountMinor": payment.AmountMinor,
		"refundedAt":  payment.UpdatedAt,
		"actorId":     actorID,
		"reason":      reason,
	})
	if err != nil {
		return err
	}
	return s.Enqueue(ctx, domain.OutboxEvent{
		ID:            s.newID("obx"),
		AggregateType: domain.OutboxAggregatePayment,
		AggregateID:   payment.ID,
		EventType:     domain.OutboxEventPaymentRefunded,
		Payload:       payload,
		CreatedAt:     s.now(),
	})
}

func (s *OutboxService) RecordFiscalDocumentCreated(ctx context.Context, document domain.FiscalDocument, storeID string) error {
	payload, err := json.Marshal(map[string]any{
		"storeId":         storeID,
		"fiscalDocumentId": document.ID,
		"receiptId":       document.ReceiptID,
		"kind":            document.Kind,
		"amountMinor":     document.AmountMinor,
		"deviceId":        document.DeviceID,
		"fiscalSign":      document.FiscalSign,
		"fiscalizedAt":    document.FiscalizedAt,
	})
	if err != nil {
		return err
	}
	return s.Enqueue(ctx, domain.OutboxEvent{
		ID:            s.newID("obx"),
		AggregateType: domain.OutboxAggregateFiscalDocument,
		AggregateID:   document.ID,
		EventType:     domain.OutboxEventFiscalDocumentCreated,
		Payload:       payload,
		CreatedAt:     s.now(),
	})
}

func (s *OutboxService) RecordCashMovementPosted(ctx context.Context, movement domain.CashMovement) error {
	payload, err := json.Marshal(map[string]any{
		"storeId":           movement.StoreID,
		"cashMovementId":    movement.ID,
		"type":              movement.Type,
		"fromContainerId":   movement.FromContainerID,
		"fromContainerType": movement.FromContainerType,
		"toContainerId":     movement.ToContainerID,
		"toContainerType":   movement.ToContainerType,
		"amountMinor":       movement.AmountMinor,
		"currency":          movement.Currency,
		"actorId":           movement.ActorID,
		"postedAt":          movement.CreatedAt,
	})
	if err != nil {
		return err
	}
	return s.Enqueue(ctx, domain.OutboxEvent{
		ID:            s.newID("obx"),
		AggregateType: domain.OutboxAggregateCashMovement,
		AggregateID:   movement.ID,
		EventType:     domain.OutboxEventCashMovementPosted,
		Payload:       payload,
		CreatedAt:     s.now(),
	})
}

func (s *OutboxService) RecordOperationalDayClosed(ctx context.Context, day domain.OperationalDay) error {
	payload, err := json.Marshal(map[string]any{
		"storeId":      day.StoreID,
		"operationalDayId": day.ID,
		"businessDate": day.BusinessDate,
		"closedById":   day.ClosedByID,
		"closedAt":     day.ClosedAt,
	})
	if err != nil {
		return err
	}
	return s.Enqueue(ctx, domain.OutboxEvent{
		ID:            s.newID("obx"),
		AggregateType: domain.OutboxAggregateOperationalDay,
		AggregateID:   day.ID,
		EventType:     domain.OutboxEventOperationalDayClosed,
		Payload:       payload,
		CreatedAt:     s.now(),
	})
}

func (s *OutboxService) RecordReturnSettled(ctx context.Context, ret domain.Return, paymentIDs []string, storeID string, actorID string) error {
	payload, err := json.Marshal(map[string]any{
		"storeId":     storeID,
		"returnId":    ret.ID,
		"receiptId":   ret.ReceiptID,
		"totalMinor":  ret.TotalMinor,
		"paymentIds":  paymentIDs,
		"settledAt":   s.now(),
		"actorId":     actorID,
	})
	if err != nil {
		return err
	}
	return s.Enqueue(ctx, domain.OutboxEvent{
		ID:            s.newID("obx"),
		AggregateType: domain.OutboxAggregateReturn,
		AggregateID:   ret.ID,
		EventType:     domain.OutboxEventReturnSettled,
		Payload:       payload,
		CreatedAt:     s.now(),
	})
}

func recordOutbox(ctx context.Context, recorder OutboxRecorder, fn func(context.Context, OutboxRecorder) error) error {
	if recorder == nil {
		return nil
	}
	return fn(ctx, recorder)
}
