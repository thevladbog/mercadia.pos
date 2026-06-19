package memory

import (
	"context"
	"sort"

	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
)

func WithDemoActors() StoreOption {
	return func(store *Store) {
		for _, actor := range demoActors() {
			store.actors[actor.ID] = actor
		}
	}
}

func demoActors() []domain.Actor {
	return []domain.Actor{
		{ID: "cashier-1", PIN: "1234", Roles: []domain.Role{domain.RoleCashier}},
		{ID: "senior-1", PIN: "5678", Roles: []domain.Role{domain.RoleSeniorCashier}},
		{ID: "admin-1", PIN: "9999", Roles: []domain.Role{domain.RoleAdmin}},
	}
}

func (s *Store) FindActor(ctx context.Context, actorID string) (domain.Actor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	actor, ok := s.actors[actorID]
	if !ok {
		return domain.Actor{}, app.ErrActorNotFound
	}
	return actor, nil
}

func (s *Store) SaveSession(ctx context.Context, session domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.Token] = session
	return nil
}

func (s *Store) FindSessionByToken(ctx context.Context, token string) (domain.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[token]
	if !ok {
		return domain.Session{}, app.ErrSessionNotFound
	}
	return session, nil
}

func (s *Store) SaveReturn(ctx context.Context, ret domain.Return) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.returns[ret.ID]; !exists {
		s.returnsByStore[ret.StoreID] = append(s.returnsByStore[ret.StoreID], ret.ID)
	}
	s.returns[ret.ID] = cloneReturn(ret)
	return nil
}

func (s *Store) FindReturn(ctx context.Context, returnID string) (domain.Return, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ret, ok := s.returns[returnID]
	if !ok {
		return domain.Return{}, app.ErrReturnNotFound
	}
	return cloneReturn(ret), nil
}

func (s *Store) ListReturnsByReceipt(ctx context.Context, receiptID string) ([]domain.Return, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.Return, 0)
	for _, ret := range s.returns {
		if ret.ReceiptID == receiptID {
			result = append(result, cloneReturn(ret))
		}
	}
	return result, nil
}

func (s *Store) SaveOperationJournalEntry(ctx context.Context, entry domain.OperationJournalEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.journalEntries[entry.ID]; !exists {
		s.journalByStore[entry.StoreID] = append(s.journalByStore[entry.StoreID], entry.ID)
	}
	s.journalEntries[entry.ID] = entry
	return nil
}

func (s *Store) ListOperationJournalEntries(ctx context.Context, storeID string, params app.PageParams) (app.PageResult[domain.OperationJournalEntry], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entryIDs := s.journalByStore[storeID]
	entries := make([]domain.OperationJournalEntry, 0, len(entryIDs))
	for _, entryID := range entryIDs {
		entry, ok := s.journalEntries[entryID]
		if ok {
			entries = append(entries, entry)
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].CreatedAt.After(entries[j].CreatedAt)
	})
	return app.PaginateSlice(entries, params), nil
}

func cloneReturn(ret domain.Return) domain.Return {
	ret.Lines = append([]domain.ReturnLine(nil), ret.Lines...)
	return ret
}
