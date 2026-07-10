package domain

import (
	"errors"
	"time"
)

type ChangeFundRequestStatus string

const (
	ChangeFundRequestStatusRequested ChangeFundRequestStatus = "requested"
	ChangeFundRequestStatusFulfilled ChangeFundRequestStatus = "fulfilled"
)

var (
	ErrInvalidChangeFundRequestInput     = errors.New("invalid change fund request input")
	ErrChangeFundRequestAlreadyFulfilled = errors.New("change fund request is already fulfilled")
)

type ChangeFundRequest struct {
	ID             string
	StoreID        string
	ShiftID        string
	ActorID        string // the requesting cashier
	AmountMinor    int64
	Currency       string
	Reason         string
	Breakdown      *DenominationBreakdown // cashier-typed at the POS; carried through unedited
	Status         ChangeFundRequestStatus
	FulfilledByID  string // empty until fulfilled
	SafeID         string // the safe that funded it; empty until fulfilled
	CashMovementID string // links to the resulting CashMovement; empty until fulfilled
	CreatedAt      time.Time
	FulfilledAt    time.Time
}

type CreateChangeFundRequestInput struct {
	ID          string
	StoreID     string
	ShiftID     string
	ActorID     string
	AmountMinor int64
	Currency    string
	Reason      string
	Breakdown   *DenominationBreakdown
	Now         time.Time
}

func NewChangeFundRequest(input CreateChangeFundRequestInput) (ChangeFundRequest, error) {
	if input.ID == "" || input.StoreID == "" || input.ShiftID == "" || input.ActorID == "" || input.AmountMinor <= 0 {
		return ChangeFundRequest{}, ErrInvalidChangeFundRequestInput
	}
	if input.Currency == "" {
		input.Currency = "RUB"
	}
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	}
	if err := validateDenominationBreakdown(input.Breakdown, input.AmountMinor); err != nil {
		return ChangeFundRequest{}, err
	}

	return ChangeFundRequest{
		ID:          input.ID,
		StoreID:     input.StoreID,
		ShiftID:     input.ShiftID,
		ActorID:     input.ActorID,
		AmountMinor: input.AmountMinor,
		Currency:    input.Currency,
		Reason:      input.Reason,
		Breakdown:   input.Breakdown,
		Status:      ChangeFundRequestStatusRequested,
		CreatedAt:   input.Now,
	}, nil
}

func (r *ChangeFundRequest) Fulfill(fulfilledByID string, safeID string, cashMovementID string, now time.Time) error {
	if r.Status != ChangeFundRequestStatusRequested {
		return ErrChangeFundRequestAlreadyFulfilled
	}
	if fulfilledByID == "" || fulfilledByID == r.ActorID || safeID == "" || cashMovementID == "" {
		return ErrInvalidChangeFundRequestInput
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	r.Status = ChangeFundRequestStatusFulfilled
	r.FulfilledByID = fulfilledByID
	r.SafeID = safeID
	r.CashMovementID = cashMovementID
	r.FulfilledAt = now
	return nil
}
