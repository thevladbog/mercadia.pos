package domain

import (
	"errors"
	"time"
)

var ErrInvalidCentralSessionInput = errors.New("invalid central session input")

type CentralSession struct {
	Token     string
	UserID    string
	Roles     []CentralRole
	CreatedAt time.Time
	ExpiresAt time.Time
}

func NewCentralSession(user CentralUser, token string, now time.Time, ttl time.Duration) (CentralSession, error) {
	if token == "" || user.ID == "" {
		return CentralSession{}, ErrInvalidCentralSessionInput
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if ttl <= 0 {
		ttl = 12 * time.Hour
	}

	roles := append([]CentralRole(nil), user.Roles...)
	return CentralSession{
		Token:     token,
		UserID:    user.ID,
		Roles:     roles,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}, nil
}

func (s CentralSession) IsExpired(now time.Time) bool {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return !now.Before(s.ExpiresAt)
}
