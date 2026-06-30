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

type AuthService struct {
	actors             ActorRepository
	sessions           SessionRepository
	credentialPolicies StoreCredentialPolicyRepository
	now                func() time.Time
	newToken           func() string
}

func NewAuthService(actors ActorRepository, sessions SessionRepository) *AuthService {
	credentialPolicies, _ := actors.(StoreCredentialPolicyRepository)
	return &AuthService{
		actors:             actors,
		sessions:           sessions,
		credentialPolicies: credentialPolicies,
		now: func() time.Time {
			return time.Now().UTC()
		},
		newToken: newSessionToken,
	}
}

type CreateSessionCommand struct {
	ActorID          string
	PIN              string
	StoreID          string
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
	credentialFactor, err := s.validateCredentialFactor(ctx, actor, command.StoreID, command.CredentialFactor)
	if err != nil {
		return SessionResult{}, err
	}

	now := s.now()
	session, err := domain.NewSession(actor, s.newToken(), now, 12*time.Hour, credentialFactor)
	if err != nil {
		return SessionResult{}, err
	}
	if err := s.sessions.SaveSession(ctx, session); err != nil {
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
