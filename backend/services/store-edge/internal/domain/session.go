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

var (
	ErrInvalidSessionInput = errors.New("invalid session input")
	ErrInvalidActorInput   = errors.New("invalid actor input")
	ErrInvalidPIN          = errors.New("invalid pin")
)

type Actor struct {
	ID    string
	PIN   string
	Roles []Role
}

type Session struct {
	Token     string
	ActorID   string
	Roles     []Role
	CreatedAt time.Time
	ExpiresAt time.Time
}

type CreateSessionInput struct {
	ActorID string
	PIN     string
	Now     time.Time
	TTL     time.Duration
}

func NewSession(actor Actor, token string, now time.Time, ttl time.Duration) (Session, error) {
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
	return Session{
		Token:     token,
		ActorID:   actor.ID,
		Roles:     roles,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
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
