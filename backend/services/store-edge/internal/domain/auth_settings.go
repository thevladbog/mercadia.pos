package domain

import (
	"errors"
	"time"
)

const (
	DefaultFailedAttemptLimit     = 5
	DefaultLockoutDurationSeconds = 900
	DefaultPOSAutoLockSeconds     = 300
	MinFailedAttemptLimit         = 1
	MaxFailedAttemptLimit         = 20
	MinLockoutDurationSeconds     = 60
	MaxLockoutDurationSeconds     = 86400
	MinPOSAutoLockSeconds         = 30
	MaxPOSAutoLockSeconds         = 86400
)

var ErrInvalidStoreAuthSettingsInput = errors.New("invalid store auth settings input")

type StoreAuthSettings struct {
	StoreID                string
	FailedAttemptLimit     int
	LockoutDurationSeconds int
	POSAutoLockSeconds     int
	UpdatedByID            string
	UpdatedAt              time.Time
}

type CreateStoreAuthSettingsInput struct {
	StoreID                string
	FailedAttemptLimit     int
	LockoutDurationSeconds int
	POSAutoLockSeconds     int
	UpdatedByID            string
	Now                    time.Time
}

func DefaultStoreAuthSettings(storeID string) StoreAuthSettings {
	return StoreAuthSettings{
		StoreID:                storeID,
		FailedAttemptLimit:     DefaultFailedAttemptLimit,
		LockoutDurationSeconds: DefaultLockoutDurationSeconds,
		POSAutoLockSeconds:     DefaultPOSAutoLockSeconds,
	}
}

func NewStoreAuthSettings(input CreateStoreAuthSettingsInput) (StoreAuthSettings, error) {
	if input.StoreID == "" || input.UpdatedByID == "" ||
		input.FailedAttemptLimit < MinFailedAttemptLimit || input.FailedAttemptLimit > MaxFailedAttemptLimit ||
		input.LockoutDurationSeconds < MinLockoutDurationSeconds || input.LockoutDurationSeconds > MaxLockoutDurationSeconds ||
		input.POSAutoLockSeconds < MinPOSAutoLockSeconds || input.POSAutoLockSeconds > MaxPOSAutoLockSeconds {
		return StoreAuthSettings{}, ErrInvalidStoreAuthSettingsInput
	}
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	}
	return StoreAuthSettings{
		StoreID:                input.StoreID,
		FailedAttemptLimit:     input.FailedAttemptLimit,
		LockoutDurationSeconds: input.LockoutDurationSeconds,
		POSAutoLockSeconds:     input.POSAutoLockSeconds,
		UpdatedByID:            input.UpdatedByID,
		UpdatedAt:              input.Now,
	}, nil
}

type AuthAttempt struct {
	ID                    string
	StoreID               string
	ActorID               string
	TerminalID            string
	CredentialKind        CredentialKind
	CredentialFingerprint string
	Successful            bool
	FailureReason         string
	CreatedAt             time.Time
}

type CreateAuthAttemptInput struct {
	ID                    string
	StoreID               string
	ActorID               string
	TerminalID            string
	CredentialKind        CredentialKind
	CredentialFingerprint string
	Successful            bool
	FailureReason         string
	Now                   time.Time
}

func NewAuthAttempt(input CreateAuthAttemptInput) (AuthAttempt, error) {
	if input.ID == "" || input.StoreID == "" || input.ActorID == "" || (!input.Successful && input.FailureReason == "") {
		return AuthAttempt{}, ErrInvalidSessionInput
	}
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	}
	return AuthAttempt{
		ID:                    input.ID,
		StoreID:               input.StoreID,
		ActorID:               input.ActorID,
		TerminalID:            input.TerminalID,
		CredentialKind:        input.CredentialKind,
		CredentialFingerprint: input.CredentialFingerprint,
		Successful:            input.Successful,
		FailureReason:         input.FailureReason,
		CreatedAt:             input.Now,
	}, nil
}
