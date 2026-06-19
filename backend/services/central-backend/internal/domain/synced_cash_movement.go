package domain

import (
	"errors"
	"time"
)

var ErrInvalidSyncedCashMovementInput = errors.New("invalid synced cash movement input")

type SyncedCashMovement struct {
	ID                string
	StoreID           string
	Type              string
	FromContainerID   string
	FromContainerType string
	ToContainerID     string
	ToContainerType   string
	AmountMinor       int64
	Currency          string
	ActorID           string
	PostedAt          time.Time
	SourceEventID     string
	SyncedAt          time.Time
}

func NewSyncedCashMovement(movement SyncedCashMovement) (SyncedCashMovement, error) {
	if movement.ID == "" || movement.StoreID == "" || movement.Type == "" ||
		movement.FromContainerID == "" || movement.FromContainerType == "" ||
		movement.ToContainerID == "" || movement.ToContainerType == "" ||
		movement.AmountMinor <= 0 || movement.ActorID == "" || movement.SourceEventID == "" {
		return SyncedCashMovement{}, ErrInvalidSyncedCashMovementInput
	}
	if movement.PostedAt.IsZero() {
		return SyncedCashMovement{}, ErrInvalidSyncedCashMovementInput
	}
	return movement, nil
}
