package app

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/domain"
)

var (
	ErrInvalidAuthCommand = errors.New("invalid auth command")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrCentralUserNotFound = errors.New("central user not found")
)

type CentralUserRepository interface {
	SaveUser(ctx context.Context, user domain.CentralUser) error
	FindUser(ctx context.Context, userID string) (domain.CentralUser, error)
	FindUserByEmail(ctx context.Context, email string) (domain.CentralUser, error)
	ListUsers(ctx context.Context) ([]domain.CentralUser, error)
}

type CentralSessionRepository interface {
	SaveSession(ctx context.Context, session domain.CentralSession) error
	FindSessionByToken(ctx context.Context, token string) (domain.CentralSession, error)
}

type AuthService struct {
	users    CentralUserRepository
	sessions CentralSessionRepository
	now      func() time.Time
	newToken func() string
}

func NewAuthService(users CentralUserRepository, sessions CentralSessionRepository) *AuthService {
	return &AuthService{
		users:    users,
		sessions: sessions,
		now: func() time.Time {
			return time.Now().UTC()
		},
		newToken: newSessionToken,
	}
}

type CreateSessionCommand struct {
	Email    string
	Password string
}

type SessionResult struct {
	Token     string
	UserID    string
	Roles     []domain.CentralRole
	ExpiresAt time.Time
}

func (s *AuthService) CreateSession(ctx context.Context, command CreateSessionCommand) (SessionResult, error) {
	if command.Email == "" || command.Password == "" {
		return SessionResult{}, ErrInvalidAuthCommand
	}

	user, err := s.users.FindUserByEmail(ctx, command.Email)
	if err != nil {
		if errors.Is(err, ErrCentralUserNotFound) {
			return SessionResult{}, ErrInvalidCredentials
		}
		return SessionResult{}, err
	}
	if !user.Active {
		return SessionResult{}, ErrInvalidCredentials
	}
	if err := CheckPassword(user.PasswordHash, command.Password); err != nil {
		return SessionResult{}, ErrInvalidCredentials
	}

	now := s.now()
	session, err := domain.NewCentralSession(user, s.newToken(), now, 12*time.Hour)
	if err != nil {
		return SessionResult{}, err
	}
	if err := s.sessions.SaveSession(ctx, session); err != nil {
		return SessionResult{}, err
	}

	return SessionResult{
		Token:     session.Token,
		UserID:    session.UserID,
		Roles:     append([]domain.CentralRole(nil), session.Roles...),
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
		UserID:    session.UserID,
		Roles:     append([]domain.CentralRole(nil), session.Roles...),
		ExpiresAt: session.ExpiresAt,
	}, nil
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
