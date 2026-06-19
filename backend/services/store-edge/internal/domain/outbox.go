package domain

import (
	"encoding/json"
	"errors"
	"time"
)

var ErrInvalidOutboxEvent = errors.New("invalid outbox event")

const (
	OutboxAggregatePayment        = "payment"
	OutboxAggregateFiscalDocument = "fiscal_document"
	OutboxAggregateCashMovement   = "cash_movement"
	OutboxAggregateOperationalDay = "operational_day"
	OutboxAggregateReturn         = "return"
)

const (
	OutboxEventPaymentCaptured       = "payment.captured"
	OutboxEventPaymentCancelled      = "payment.cancelled"
	OutboxEventPaymentRefunded       = "payment.refunded"
	OutboxEventFiscalDocumentCreated = "fiscal.document.created"
	OutboxEventCashMovementPosted    = "cash.movement.posted"
	OutboxEventOperationalDayClosed  = "operational_day.closed"
	OutboxEventReturnSettled         = "return.settled"
)

type OutboxEvent struct {
	ID            string
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       json.RawMessage
	CreatedAt     time.Time
	PublishedAt   *time.Time
}

func NewOutboxEvent(event OutboxEvent) (OutboxEvent, error) {
	if event.ID == "" || event.AggregateType == "" || event.AggregateID == "" || event.EventType == "" {
		return OutboxEvent{}, ErrInvalidOutboxEvent
	}
	if len(event.Payload) == 0 {
		event.Payload = json.RawMessage(`{}`)
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	return event, nil
}

func (e OutboxEvent) IsPublished() bool {
	return e.PublishedAt != nil
}
