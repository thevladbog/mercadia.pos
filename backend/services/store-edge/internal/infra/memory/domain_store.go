package memory

import (
	"context"
	"sort"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
)

func WithDemoActors() StoreOption {
	return func(store *Store) {
		for _, actor := range demoActors() {
			store.actors[actor.ID] = actor
		}
		store.credentialPolicies["store-1"] = domain.CredentialPolicy{
			Required: true,
			AllowedKinds: []domain.CredentialKind{
				domain.CredentialKindIButton,
				domain.CredentialKindMSRCard,
				domain.CredentialKindBarcodeCard,
			},
		}
	}
}

func demoActors() []domain.Actor {
	notRequired := domain.CredentialPolicy{Required: false}
	return []domain.Actor{
		{ID: "cashier-1", PIN: "1234", Roles: []domain.Role{domain.RoleCashier}, CredentialPolicy: &notRequired},
		{
			ID:    "senior-1",
			PIN:   "5678",
			Roles: []domain.Role{domain.RoleSeniorCashier},
			CredentialBindings: []domain.CredentialBinding{
				{Kind: domain.CredentialKindIButton, TokenHash: app.HashCredentialToken("demo-ibutton-senior-1"), MaskedToken: "iButton demo ****0001", Active: true},           // #nosec G101 -- deterministic demo binding fixture, not a secret.
				{Kind: domain.CredentialKindMSRCard, TokenHash: app.HashCredentialToken("demo-msr-senior-1"), MaskedToken: "MSR staff demo ****0001", Active: true},             // #nosec G101 -- deterministic demo binding fixture, not a secret.
				{Kind: domain.CredentialKindBarcodeCard, TokenHash: app.HashCredentialToken("demo-barcode-senior-1"), MaskedToken: "Barcode staff demo ****0001", Active: true}, // #nosec G101 -- deterministic demo binding fixture, not a secret.
			},
		},
		{ID: "admin-1", PIN: "9999", Roles: []domain.Role{domain.RoleAdmin}, CredentialPolicy: &notRequired},
	}
}

func (s *Store) FindActor(ctx context.Context, actorID string) (domain.Actor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	actor, ok := s.actors[actorID]
	if !ok {
		return domain.Actor{}, app.ErrActorNotFound
	}
	return cloneActor(actor), nil
}

func (s *Store) ListActors(ctx context.Context) ([]domain.Actor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.actors))
	for id := range s.actors {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	actors := make([]domain.Actor, 0, len(ids))
	for _, id := range ids {
		actors = append(actors, cloneActor(s.actors[id]))
	}
	return actors, nil
}

func (s *Store) SaveActor(ctx context.Context, actor domain.Actor) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.actors[actor.ID]; !ok {
		return app.ErrActorNotFound
	}
	s.actors[actor.ID] = cloneActor(actor)
	return nil
}

func (s *Store) UpdateActorCredentialPolicy(ctx context.Context, actorID string, policy *domain.CredentialPolicy) (domain.Actor, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	actor, ok := s.actors[actorID]
	if !ok {
		return domain.Actor{}, app.ErrActorNotFound
	}
	actor.CredentialPolicy = cloneCredentialPolicyPointer(policy)
	s.actors[actorID] = cloneActor(actor)
	return cloneActor(actor), nil
}

func (s *Store) UpdateActorCredentialBindings(ctx context.Context, actorID string, update func([]domain.CredentialBinding) ([]domain.CredentialBinding, error)) (domain.Actor, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	actor, ok := s.actors[actorID]
	if !ok {
		return domain.Actor{}, app.ErrActorNotFound
	}
	bindings := append([]domain.CredentialBinding(nil), actor.CredentialBindings...)
	updated, err := update(bindings)
	if err != nil {
		return domain.Actor{}, err
	}
	actor.CredentialBindings = append([]domain.CredentialBinding(nil), updated...)
	s.actors[actorID] = cloneActor(actor)
	return cloneActor(actor), nil
}

func (s *Store) FindStoreCredentialPolicy(ctx context.Context, storeID string) (domain.CredentialPolicy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policy, ok := s.credentialPolicies[storeID]
	if !ok {
		return domain.CredentialPolicy{}, nil
	}
	policy.AllowedKinds = append([]domain.CredentialKind(nil), policy.AllowedKinds...)
	return policy, nil
}

func (s *Store) SaveStoreCredentialPolicy(ctx context.Context, storeID string, policy domain.CredentialPolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	policy.AllowedKinds = append([]domain.CredentialKind(nil), policy.AllowedKinds...)
	s.credentialPolicies[storeID] = policy
	return nil
}

func (s *Store) FindStoreAuthSettings(ctx context.Context, storeID string) (domain.StoreAuthSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	settings, ok := s.authSettings[storeID]
	if !ok {
		return domain.DefaultStoreAuthSettings(storeID), nil
	}
	return settings, nil
}

func (s *Store) SaveStoreAuthSettings(ctx context.Context, settings domain.StoreAuthSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.authSettings[settings.StoreID] = settings
	return nil
}

func (s *Store) SaveAuthAttempt(ctx context.Context, attempt domain.AuthAttempt) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.authAttempts = append(s.authAttempts, attempt)
	return nil
}

func (s *Store) ListAuthAttempts(ctx context.Context, filter app.AuthAttemptFilter, params app.PageParams) (app.PageResult[domain.AuthAttempt], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	attempts := make([]domain.AuthAttempt, 0, len(s.authAttempts))
	for _, attempt := range s.authAttempts {
		if filter.StoreID != "" && attempt.StoreID != filter.StoreID {
			continue
		}
		if filter.ActorID != "" && attempt.ActorID != filter.ActorID {
			continue
		}
		if filter.TerminalID != "" && attempt.TerminalID != filter.TerminalID {
			continue
		}
		if filter.Successful != nil && attempt.Successful != *filter.Successful {
			continue
		}
		if !filter.Since.IsZero() && attempt.CreatedAt.Before(filter.Since) {
			continue
		}
		if !filter.Until.IsZero() && attempt.CreatedAt.After(filter.Until) {
			continue
		}
		attempts = append(attempts, attempt)
	}
	sort.Slice(attempts, func(i, j int) bool {
		if attempts[i].CreatedAt.Equal(attempts[j].CreatedAt) {
			return attempts[i].ID > attempts[j].ID
		}
		return attempts[i].CreatedAt.After(attempts[j].CreatedAt)
	})
	return app.PaginateSlice(attempts, params), nil
}

func (s *Store) CountFailedAuthAttemptsSinceLastSuccess(ctx context.Context, storeID string, actorID string, since time.Time) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lastSuccessAt := since
	foundSuccess := false
	for _, attempt := range s.authAttempts {
		if attempt.StoreID == storeID && attempt.ActorID == actorID && attempt.Successful && !attempt.CreatedAt.Before(since) && (!foundSuccess || attempt.CreatedAt.After(lastSuccessAt)) {
			lastSuccessAt = attempt.CreatedAt
			foundSuccess = true
		}
	}
	count := 0
	for _, attempt := range s.authAttempts {
		if attempt.StoreID != storeID || attempt.ActorID != actorID || attempt.Successful || attempt.FailureReason == "locked" {
			continue
		}
		if foundSuccess && !attempt.CreatedAt.After(lastSuccessAt) {
			continue
		}
		if !foundSuccess && attempt.CreatedAt.Before(lastSuccessAt) {
			continue
		}
		count++
	}
	return count, nil
}

func cloneActor(actor domain.Actor) domain.Actor {
	actor.Roles = append([]domain.Role(nil), actor.Roles...)
	actor.CredentialPolicy = cloneCredentialPolicyPointer(actor.CredentialPolicy)
	actor.CredentialBindings = append([]domain.CredentialBinding(nil), actor.CredentialBindings...)
	return actor
}

func cloneCredentialPolicyPointer(policy *domain.CredentialPolicy) *domain.CredentialPolicy {
	if policy == nil {
		return nil
	}
	cloned := *policy
	cloned.AllowedKinds = append([]domain.CredentialKind(nil), policy.AllowedKinds...)
	return &cloned
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

func (s *Store) ListReturnsByStore(ctx context.Context, storeID string) ([]domain.Return, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.Return, 0)
	for _, returnID := range s.returnsByStore[storeID] {
		ret, ok := s.returns[returnID]
		if !ok {
			continue
		}
		result = append(result, cloneReturn(ret))
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
