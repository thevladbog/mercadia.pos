package domain

import (
	"errors"
	"time"
)

type CashRecountStatus string
type CashRecountResolutionStatus string

const (
	CashRecountStatusBalanced    CashRecountStatus = "balanced"
	CashRecountStatusDiscrepancy CashRecountStatus = "discrepancy"

	CashRecountResolutionStatusNotRequired CashRecountResolutionStatus = "not_required"
	CashRecountResolutionStatusOpen        CashRecountResolutionStatus = "open"
	CashRecountResolutionStatusResolved    CashRecountResolutionStatus = "resolved"
)

var (
	ErrInvalidCashRecountInput        = errors.New("invalid cash recount input")
	ErrCashRecountResolutionNotNeeded = errors.New("cash recount resolution is not needed")
	ErrCashRecountAlreadyResolved     = errors.New("cash recount already resolved")
)

type CashRecount struct {
	ID               string
	StoreID          string
	ContainerID      string
	ContainerType    CashContainerType
	Currency         string
	ExpectedMinor    int64
	CountedMinor     int64
	DiscrepancyMinor int64
	Reason           string
	ActorID          string
	ApprovedByID     string
	Status           CashRecountStatus
	ResolutionStatus CashRecountResolutionStatus
	ResolutionNote   string
	ResolvedByID     string
	ResolvedAt       time.Time
	CreatedAt        time.Time
}

type CreateCashRecountInput struct {
	ID            string
	StoreID       string
	ContainerID   string
	ContainerType CashContainerType
	Currency      string
	ExpectedMinor int64
	CountedMinor  int64
	Reason        string
	ActorID       string
	ApprovedByID  string
	Now           time.Time
}

func CreateCashRecount(input CreateCashRecountInput) (CashRecount, error) {
	if input.ID == "" || input.StoreID == "" || input.ContainerID == "" ||
		input.ContainerType == "" || input.CountedMinor < 0 || input.ActorID == "" {
		return CashRecount{}, ErrInvalidCashRecountInput
	}
	if input.Currency == "" {
		input.Currency = "RUB"
	}
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	}

	discrepancy := input.CountedMinor - input.ExpectedMinor
	status := CashRecountStatusBalanced
	resolutionStatus := CashRecountResolutionStatusNotRequired
	if discrepancy != 0 {
		status = CashRecountStatusDiscrepancy
		resolutionStatus = CashRecountResolutionStatusOpen
	}

	return CashRecount{
		ID:               input.ID,
		StoreID:          input.StoreID,
		ContainerID:      input.ContainerID,
		ContainerType:    input.ContainerType,
		Currency:         input.Currency,
		ExpectedMinor:    input.ExpectedMinor,
		CountedMinor:     input.CountedMinor,
		DiscrepancyMinor: discrepancy,
		Reason:           input.Reason,
		ActorID:          input.ActorID,
		ApprovedByID:     input.ApprovedByID,
		Status:           status,
		ResolutionStatus: resolutionStatus,
		CreatedAt:        input.Now,
	}, nil
}

func (r *CashRecount) Resolve(resolutionNote string, resolvedByID string, now time.Time) error {
	if r.Status != CashRecountStatusDiscrepancy {
		return ErrCashRecountResolutionNotNeeded
	}
	if r.ResolutionStatus == CashRecountResolutionStatusResolved {
		return ErrCashRecountAlreadyResolved
	}
	if resolutionNote == "" || resolvedByID == "" {
		return ErrInvalidCashRecountInput
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	r.ResolutionStatus = CashRecountResolutionStatusResolved
	r.ResolutionNote = resolutionNote
	r.ResolvedByID = resolvedByID
	r.ResolvedAt = now
	return nil
}
