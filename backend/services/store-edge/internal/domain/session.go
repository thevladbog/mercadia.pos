package domain

import (
	"errors"
	"time"
)

type Role string

const (
	RoleCashier       Role = "cashier"
	RoleSeniorCashier Role = "senior_cashier"
	RoleAdmin         Role = "admin"
)

type CredentialKind string

const (
	CredentialKindIButton     CredentialKind = "ibutton"
	CredentialKindMSRCard     CredentialKind = "msr_card"
	CredentialKindBarcodeCard CredentialKind = "barcode_card"
)

var (
	ErrInvalidSessionInput = errors.New("invalid session input")
	ErrInvalidActorInput   = errors.New("invalid actor input")
	ErrInvalidPIN          = errors.New("invalid pin")
)

type Actor struct {
	ID                 string
	PIN                string
	Roles              []Role
	CredentialPolicy   *CredentialPolicy
	CredentialBindings []CredentialBinding
}

type CredentialPolicy struct {
	Required     bool
	AllowedKinds []CredentialKind
}

type CredentialBinding struct {
	Kind        CredentialKind
	TokenHash   string
	MaskedToken string
	Active      bool
}

type SubmittedCredentialFactor struct {
	Kind      CredentialKind
	Token     string
	DeviceID  string
	CommandID string
}

type SessionCredentialFactor struct {
	Kind             CredentialKind
	DeviceID         string
	CommandID        string
	TokenFingerprint string
	MaskedToken      string
}

type Session struct {
	Token            string
	ActorID          string
	Roles            []Role
	CredentialFactor *SessionCredentialFactor
	CreatedAt        time.Time
	ExpiresAt        time.Time
}

type CreateSessionInput struct {
	ActorID string
	PIN     string
	Now     time.Time
	TTL     time.Duration
}

func NewSession(actor Actor, token string, now time.Time, ttl time.Duration, credentialFactor *SessionCredentialFactor) (Session, error) {
	if token == "" || actor.ID == "" {
		return Session{}, ErrInvalidSessionInput
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if ttl <= 0 {
		ttl = 12 * time.Hour
	}

	roles := append([]Role(nil), actor.Roles...)
	var factor *SessionCredentialFactor
	if credentialFactor != nil {
		cloned := *credentialFactor
		factor = &cloned
	}
	return Session{
		Token:            token,
		ActorID:          actor.ID,
		Roles:            roles,
		CredentialFactor: factor,
		CreatedAt:        now,
		ExpiresAt:        now.Add(ttl),
	}, nil
}

func (s Session) IsExpired(now time.Time) bool {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return !now.Before(s.ExpiresAt)
}

func (a Actor) HasRole(role Role) bool {
	for _, candidate := range a.Roles {
		if candidate == role {
			return true
		}
	}
	return false
}
