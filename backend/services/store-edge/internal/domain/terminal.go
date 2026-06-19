package domain

import (
	"errors"
	"time"
)

type TerminalKind string
type TerminalStatus string

const (
	TerminalKindPOS              TerminalKind   = "pos"
	TerminalKindSCO              TerminalKind   = "sco"
	TerminalKindSeniorCashier    TerminalKind   = "senior_cashier"
	TerminalKindSeniorCashierWeb TerminalKind   = "senior_cashier_web"
	TerminalKindAssistantStation TerminalKind   = "assistant_station"
	TerminalKindStoreAdmin       TerminalKind   = "store_admin"
	TerminalStatusOnline         TerminalStatus = "online"
)

var ErrInvalidTerminalInput = errors.New("invalid terminal input")

type Terminal struct {
	ID              string
	StoreID         string
	Kind            TerminalKind
	Status          TerminalStatus
	SoftwareVersion string
	LastSeenAt      time.Time
	UpdatedAt       time.Time
}

type RecordTerminalHeartbeatInput struct {
	ID              string
	StoreID         string
	Kind            TerminalKind
	SoftwareVersion string
	Now             time.Time
}

func RecordTerminalHeartbeat(input RecordTerminalHeartbeatInput) (Terminal, error) {
	if input.ID == "" || input.StoreID == "" || input.Kind == "" {
		return Terminal{}, ErrInvalidTerminalInput
	}
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	}

	return Terminal{
		ID:              input.ID,
		StoreID:         input.StoreID,
		Kind:            input.Kind,
		Status:          TerminalStatusOnline,
		SoftwareVersion: input.SoftwareVersion,
		LastSeenAt:      input.Now,
		UpdatedAt:       input.Now,
	}, nil
}
