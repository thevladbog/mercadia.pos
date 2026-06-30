package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

var (
	ErrInvalidAuthCommand = errors.New("invalid auth command")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAuthLocked         = errors.New("auth locked")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
	ErrPermissionDenied   = errors.New("permission denied")
)

type ActorRepository interface {
	FindActor(ctx context.Context, actorID string) (domain.Actor, error)
}

type SessionRepository interface {
	SaveSession(ctx context.Context, session domain.Session) error
	FindSessionByToken(ctx context.Context, token string) (domain.Session, error)
}

type StoreCredentialPolicyRepository interface {
	FindStoreCredentialPolicy(ctx context.Context, storeID string) (domain.CredentialPolicy, error)
}

type AuthSettingsReader interface {
	FindStoreAuthSettings(ctx context.Context, storeID string) (domain.StoreAuthSettings, error)
}

type AuthAttemptRepository interface {
	SaveAuthAttempt(ctx context.Context, attempt domain.AuthAttempt) error
	CountFailedAuthAttemptsSinceLastSuccess(ctx context.Context, storeID string, actorID string, since time.Time) (int, error)
}

type AuthService struct {
	actors             ActorRepository
	sessions           SessionRepository
	credentialPolicies StoreCredentialPolicyRepository
	authSettings       AuthSettingsReader
	authAttempts       AuthAttemptRepository
	transactions       TransactionRunner
	now                func() time.Time
	newToken           func() string
	newAttemptID       func(prefix string) string
}

type AuthOption func(*AuthService)

func NewAuthService(actors ActorRepository, sessions SessionRepository, authSettings AuthSettingsReader, authAttempts AuthAttemptRepository, options ...AuthOption) *AuthService {
	credentialPolicies, _ := actors.(StoreCredentialPolicyRepository)
	service := &AuthService{
		actors:             actors,
		sessions:           sessions,
		credentialPolicies: credentialPolicies,
		authSettings:       authSettings,
		authAttempts:       authAttempts,
		now: func() time.Time {
			return time.Now().UTC()
		},
		newToken:     newSessionToken,
		newAttemptID: randomID,
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithAuthTransactionRunner(runner TransactionRunner) AuthOption {
	return func(service *AuthService) {
		service.transactions = runner
	}
}

type CreateSessionCommand struct {
	ActorID          string
	PIN              string
	StoreID          string
	TerminalID       string
	CredentialFactor *domain.SubmittedCredentialFactor
}

type SessionResult struct {
	Token            string
	ActorID          string
	Roles            []domain.Role
	CredentialFactor *domain.SessionCredentialFactor
	ExpiresAt        time.Time
}

func (s *AuthService) CreateSession(ctx context.Context, command CreateSessionCommand) (SessionResult, error) {
	if command.ActorID == "" || command.PIN == "" || command.StoreID == "" {
		return SessionResult{}, ErrInvalidAuthCommand
	}

	now := s.now()
	settings, err := s.storeAuthSettings(ctx, command.StoreID)
	if err != nil {
		return SessionResult{}, err
	}
	locked, err := s.isAuthLocked(ctx, settings, command.ActorID, now)
	if err != nil {
		return SessionResult{}, err
	}
	if locked {
		if err := s.recordAuthAttempt(ctx, authAttemptInput{
			StoreID:          command.StoreID,
			ActorID:          command.ActorID,
			TerminalID:       command.TerminalID,
			CredentialFactor: command.CredentialFactor,
			FailureReason:    "locked",
			Now:              now,
		}); err != nil {
			return SessionResult{}, err
		}
		return SessionResult{}, ErrAuthLocked
	}

	actor, err := s.actors.FindActor(ctx, command.ActorID)
	if err != nil {
		if errors.Is(err, ErrActorNotFound) {
			if err := s.recordFailedAuthAttempt(ctx, settings, command, now, "actor_not_found"); err != nil {
				return SessionResult{}, err
			}
			return SessionResult{}, ErrInvalidCredentials
		}
		return SessionResult{}, err
	}
	if actor.PIN != command.PIN {
		if err := s.recordFailedAuthAttempt(ctx, settings, command, now, "invalid_pin"); err != nil {
			return SessionResult{}, err
		}
		return SessionResult{}, ErrInvalidCredentials
	}
	credentialFactor, err := s.validateCredentialFactor(ctx, actor, command.StoreID, command.CredentialFactor)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			if recordErr := s.recordFailedAuthAttempt(ctx, settings, command, now, "invalid_credential"); recordErr != nil {
				return SessionResult{}, recordErr
			}
		}
		return SessionResult{}, err
	}

	session, err := domain.NewSession(actor, s.newToken(), now, 12*time.Hour, credentialFactor)
	if err != nil {
		return SessionResult{}, err
	}
	if err := RunTransaction(ctx, s.transactions, func(ctx context.Context) error {
		if err := s.recordAuthAttempt(ctx, authAttemptInput{
			StoreID:          command.StoreID,
			ActorID:          command.ActorID,
			TerminalID:       command.TerminalID,
			CredentialFactor: command.CredentialFactor,
			Successful:       true,
			Now:              now,
		}); err != nil {
			return err
		}
		return s.sessions.SaveSession(ctx, session)
	}); err != nil {
		return SessionResult{}, err
	}

	return SessionResult{
		Token:            session.Token,
		ActorID:          session.ActorID,
		Roles:            append([]domain.Role(nil), session.Roles...),
		CredentialFactor: cloneSessionCredentialFactor(session.CredentialFactor),
		ExpiresAt:        session.ExpiresAt,
	}, nil
}

func (s *AuthService) storeAuthSettings(ctx context.Context, storeID string) (domain.StoreAuthSettings, error) {
	return s.authSettings.FindStoreAuthSettings(ctx, storeID)
}

func (s *AuthService) isAuthLocked(ctx context.Context, settings domain.StoreAuthSettings, actorID string, now time.Time) (bool, error) {
	if settings.FailedAttemptLimit <= 0 || settings.LockoutDurationSeconds <= 0 {
		return false, nil
	}
	since := now.Add(-time.Duration(settings.LockoutDurationSeconds) * time.Second)
	failedCount, err := s.authAttempts.CountFailedAuthAttemptsSinceLastSuccess(ctx, settings.StoreID, actorID, since)
	if err != nil {
		return false, err
	}
	return failedCount >= settings.FailedAttemptLimit, nil
}

func (s *AuthService) recordFailedAuthAttempt(ctx context.Context, settings domain.StoreAuthSettings, command CreateSessionCommand, now time.Time, reason string) error {
	if err := s.recordAuthAttempt(ctx, authAttemptInput{
		StoreID:          command.StoreID,
		ActorID:          command.ActorID,
		TerminalID:       command.TerminalID,
		CredentialFactor: command.CredentialFactor,
		FailureReason:    reason,
		Now:              now,
	}); err != nil {
		return err
	}
	locked, err := s.isAuthLocked(ctx, settings, command.ActorID, now)
	if err != nil {
		return err
	}
	if locked {
		return ErrAuthLocked
	}
	return nil
}

type authAttemptInput struct {
	StoreID          string
	ActorID          string
	TerminalID       string
	CredentialFactor *domain.SubmittedCredentialFactor
	Successful       bool
	FailureReason    string
	Now              time.Time
}

func (s *AuthService) recordAuthAttempt(ctx context.Context, input authAttemptInput) error {
	var kind domain.CredentialKind
	fingerprint := ""
	if input.CredentialFactor != nil {
		kind = input.CredentialFactor.Kind
		if input.CredentialFactor.Token != "" {
			fingerprint = tokenFingerprint(hashCredentialToken(input.CredentialFactor.Token))
		}
	}
	attempt, err := domain.NewAuthAttempt(domain.CreateAuthAttemptInput{
		ID:                    s.newAttemptID("auth_attempt"),
		StoreID:               input.StoreID,
		ActorID:               input.ActorID,
		TerminalID:            input.TerminalID,
		CredentialKind:        kind,
		CredentialFingerprint: fingerprint,
		Successful:            input.Successful,
		FailureReason:         input.FailureReason,
		Now:                   input.Now,
	})
	if err != nil {
		return err
	}
	return s.authAttempts.SaveAuthAttempt(ctx, attempt)
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
		Token:            session.Token,
		ActorID:          session.ActorID,
		Roles:            append([]domain.Role(nil), session.Roles...),
		CredentialFactor: cloneSessionCredentialFactor(session.CredentialFactor),
		ExpiresAt:        session.ExpiresAt,
	}, nil
}

func (s *AuthService) validateCredentialFactor(ctx context.Context, actor domain.Actor, storeID string, factor *domain.SubmittedCredentialFactor) (*domain.SessionCredentialFactor, error) {
	policy, err := s.effectiveCredentialPolicy(ctx, actor, storeID)
	if err != nil {
		return nil, err
	}
	if !policy.Required && factor == nil {
		return nil, nil
	}
	token := ""
	if factor != nil {
		token = strings.TrimSpace(factor.Token)
	}
	if factor == nil || factor.Kind == "" || token == "" {
		return nil, ErrInvalidCredentials
	}
	if !credentialKindAllowed(policy, factor.Kind) {
		return nil, ErrInvalidCredentials
	}

	tokenHash := hashCredentialToken(token)
	for _, binding := range actor.CredentialBindings {
		if binding.Active && binding.Kind == factor.Kind && binding.TokenHash == tokenHash {
			return &domain.SessionCredentialFactor{
				Kind:             factor.Kind,
				DeviceID:         factor.DeviceID,
				CommandID:        factor.CommandID,
				TokenFingerprint: tokenFingerprint(tokenHash),
				MaskedToken:      binding.MaskedToken,
			}, nil
		}
	}
	return nil, ErrInvalidCredentials
}

func (s *AuthService) effectiveCredentialPolicy(ctx context.Context, actor domain.Actor, storeID string) (domain.CredentialPolicy, error) {
	if actor.CredentialPolicy != nil {
		return cloneCredentialPolicy(*actor.CredentialPolicy), nil
	}
	if s.credentialPolicies == nil {
		return domain.CredentialPolicy{}, nil
	}
	policy, err := s.credentialPolicies.FindStoreCredentialPolicy(ctx, storeID)
	if err != nil {
		return domain.CredentialPolicy{}, err
	}
	return cloneCredentialPolicy(policy), nil
}

func credentialKindAllowed(policy domain.CredentialPolicy, kind domain.CredentialKind) bool {
	if len(policy.AllowedKinds) == 0 {
		return !policy.Required
	}
	for _, allowed := range policy.AllowedKinds {
		if allowed == kind {
			return true
		}
	}
	return false
}

func HashCredentialToken(token string) string {
	return hashCredentialToken(token)
}

func hashCredentialToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func tokenFingerprint(tokenHash string) string {
	if len(tokenHash) <= 12 {
		return tokenHash
	}
	return tokenHash[:12]
}

func cloneCredentialPolicy(policy domain.CredentialPolicy) domain.CredentialPolicy {
	policy.AllowedKinds = append([]domain.CredentialKind(nil), policy.AllowedKinds...)
	return policy
}

func cloneSessionCredentialFactor(factor *domain.SessionCredentialFactor) *domain.SessionCredentialFactor {
	if factor == nil {
		return nil
	}
	cloned := *factor
	return &cloned
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
