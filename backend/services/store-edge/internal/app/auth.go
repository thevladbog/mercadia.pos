package app

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

var (
	ErrInvalidAuthCommand   = errors.New("invalid auth command")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionExpired       = errors.New("session expired")
	ErrPermissionDenied     = errors.New("permission denied")
)

type ActorRepository interface {
	FindActor(ctx context.Context, actorID string) (domain.Actor, error)
}

type SessionRepository interface {
	SaveSession(ctx context.Context, session domain.Session) error
	FindSessionByToken(ctx context.Context, token string) (domain.Session, error)
}

type AuthService struct {
	actors   ActorRepository
	sessions SessionRepository
	now      func() time.Time
	newToken func() string
}

func NewAuthService(actors ActorRepository, sessions SessionRepository) *AuthService {
	return &AuthService{
		actors:   actors,
		sessions: sessions,
		now: func() time.Time {
			return time.Now().UTC()
		},
		newToken: newSessionToken,
	}
}

type CreateSessionCommand struct {
	ActorID string
	PIN     string
}

type SessionResult struct {
	Token     string
	ActorID   string
	Roles     []domain.Role
	ExpiresAt time.Time
}

func (s *AuthService) CreateSession(ctx context.Context, command CreateSessionCommand) (SessionResult, error) {
	if command.ActorID == "" || command.PIN == "" {
		return SessionResult{}, ErrInvalidAuthCommand
	}

	actor, err := s.actors.FindActor(ctx, command.ActorID)
	if err != nil {
		if errors.Is(err, ErrActorNotFound) {
			return SessionResult{}, ErrInvalidCredentials
		}
		return SessionResult{}, err
	}
	if actor.PIN != command.PIN {
		return SessionResult{}, ErrInvalidCredentials
	}

	now := s.now()
	session, err := domain.NewSession(actor, s.newToken(), now, 12*time.Hour)
	if err != nil {
		return SessionResult{}, err
	}
	if err := s.sessions.SaveSession(ctx, session); err != nil {
		return SessionResult{}, err
	}

	return SessionResult{
		Token:     session.Token,
		ActorID:   session.ActorID,
		Roles:     append([]domain.Role(nil), session.Roles...),
		ExpiresAt: session.ExpiresAt,
	}, nil
}

func (s *AuthService) ResolveSession(ctx context.Context, token string) (SessionResult, error) {
	if token == "" {
		return SessionResult{}, ErrSessionNotFound
	}
	session, err := s.sessions.FindSessionByToken(ctx, token)
	if err != nil {
		return SessionResult{}, err
	}
	if session.IsExpired(s.now()) {
		return SessionResult{}, ErrSessionExpired
	}
	return SessionResult{
		Token:     session.Token,
		ActorID:   session.ActorID,
		Roles:     append([]domain.Role(nil), session.Roles...),
		ExpiresAt: session.ExpiresAt,
	}, nil
}

func (s *AuthService) FindActorRoles(ctx context.Context, actorID string) ([]domain.Role, error) {
	actor, err := s.actors.FindActor(ctx, actorID)
	if err != nil {
		return nil, err
	}
	return append([]domain.Role(nil), actor.Roles...), nil
}

func newSessionToken() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(fmt.Sprintf("generate session token: %v", err))
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}

var ErrActorNotFound = errors.New("actor not found")
