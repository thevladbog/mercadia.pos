package memory

import (
	"context"
	"strings"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/domain"
)

func (s *Store) SaveUser(ctx context.Context, user domain.CentralUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[user.ID] = cloneCentralUser(user)
	return nil
}

func (s *Store) FindUser(ctx context.Context, userID string) (domain.CentralUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[userID]
	if !ok {
		return domain.CentralUser{}, app.ErrCentralUserNotFound
	}
	return cloneCentralUser(user), nil
}

func (s *Store) FindUserByEmail(ctx context.Context, email string) (domain.CentralUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	normalized := strings.ToLower(strings.TrimSpace(email))
	for _, user := range s.users {
		if strings.ToLower(user.Email) == normalized {
			return cloneCentralUser(user), nil
		}
	}
	return domain.CentralUser{}, app.ErrCentralUserNotFound
}

func (s *Store) ListUsers(ctx context.Context) ([]domain.CentralUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]domain.CentralUser, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, cloneCentralUser(user))
	}
	return users, nil
}

func (s *Store) SaveSession(ctx context.Context, session domain.CentralSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.sessions == nil {
		s.sessions = map[string]domain.CentralSession{}
	}
	s.sessions[session.Token] = session
	return nil
}

func (s *Store) FindSessionByToken(ctx context.Context, token string) (domain.CentralSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[token]
	if !ok {
		return domain.CentralSession{}, app.ErrSessionNotFound
	}
	return session, nil
}

func cloneCentralUser(user domain.CentralUser) domain.CentralUser {
	clone := user
	clone.Roles = append([]domain.CentralRole(nil), user.Roles...)
	return clone
}
