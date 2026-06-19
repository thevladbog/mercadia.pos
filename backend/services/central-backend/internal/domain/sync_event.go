package domain

import (
	"encoding/json"
	"errors"
	"time"
)

var ErrInvalidSyncEventInput = errors.New("invalid sync event input")

type SyncEvent struct {
	ID            string
	StoreID       string
	EventType     string
	SourceEventID string
	Payload       json.RawMessage
	OccurredAt    time.Time
	ReceivedAt    time.Time
}

func NewSyncEvent(event SyncEvent) (SyncEvent, error) {
	if event.ID == "" || event.StoreID == "" || event.EventType == "" || event.SourceEventID == "" {
		return SyncEvent{}, ErrInvalidSyncEventInput
	}
	if len(event.Payload) == 0 {
		event.Payload = json.RawMessage(`{}`)
	}
	return event, nil
}
