package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

var ErrCredentialBindingNotFound = errors.New("credential binding not found")

const (
	setStoreCredentialPolicyOperation = "credentials.set_store_policy"
	setActorCredentialPolicyOperation = "credentials.set_actor_policy"
	addCredentialBindingOperation     = "credentials.add_binding"
	revokeCredentialBindingOperation  = "credentials.revoke_binding"
)

type CredentialManagementRepository interface {
	FindActor(ctx context.Context, actorID string) (domain.Actor, error)
	ListActors(ctx context.Context) ([]domain.Actor, error)
	UpdateActorCredentialPolicy(ctx context.Context, actorID string, policy *domain.CredentialPolicy) (domain.Actor, error)
	UpdateActorCredentialBindings(ctx context.Context, actorID string, update func([]domain.CredentialBinding) ([]domain.CredentialBinding, error)) (domain.Actor, error)
	FindStoreCredentialPolicy(ctx context.Context, storeID string) (domain.CredentialPolicy, error)
	SaveStoreCredentialPolicy(ctx context.Context, storeID string, policy domain.CredentialPolicy) error
}

type CredentialManagementService struct {
	repo         CredentialManagementRepository
	idempotency  IdempotencyStore
	transactions TransactionRunner
}

type CredentialManagementOption func(*CredentialManagementService)

func NewCredentialManagementService(repo CredentialManagementRepository, idempotency IdempotencyStore, options ...CredentialManagementOption) *CredentialManagementService {
	service := &CredentialManagementService{repo: repo, idempotency: idempotency}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithCredentialManagementTransactionRunner(runner TransactionRunner) CredentialManagementOption {
	return func(service *CredentialManagementService) {
		service.transactions = runner
	}
}

type CredentialManagementResult struct {
	StoreID     string
	StorePolicy CredentialPolicyResult
	Actors      []ActorCredentialResult
}

type ActorCredentialResult struct {
	ID                 string
	Roles              []domain.Role
	CredentialPolicy   *CredentialPolicyResult
	CredentialBindings []CredentialBindingResult
}

type CredentialPolicyResult struct {
	Required     bool
	AllowedKinds []domain.CredentialKind
}

type CredentialBindingResult struct {
	Kind             domain.CredentialKind
	TokenFingerprint string
	MaskedToken      string
	Active           bool
}

type GetCredentialManagementQuery struct {
	StoreID   string
	ManagerID string
}

type SetStoreCredentialPolicyCommand struct {
	IdempotencyKey string
	StoreID        string
	ManagerID      string
	Required       bool
	AllowedKinds   []domain.CredentialKind
}

type SetActorCredentialPolicyCommand struct {
	IdempotencyKey     string
	TargetActorID      string
	ManagerID          string
	InheritStorePolicy bool
	Required           bool
	AllowedKinds       []domain.CredentialKind
}

type AddCredentialBindingCommand struct {
	IdempotencyKey string
	TargetActorID  string
	ManagerID      string
	Kind           domain.CredentialKind
	Token          string
	MaskedToken    string
}

type RevokeCredentialBindingCommand struct {
	IdempotencyKey   string
	TargetActorID    string
	ManagerID        string
	Kind             domain.CredentialKind
	TokenFingerprint string
}

func (s *CredentialManagementService) GetCredentialManagement(ctx context.Context, query GetCredentialManagementQuery) (CredentialManagementResult, error) {
	if query.StoreID == "" || query.ManagerID == "" {
		return CredentialManagementResult{}, ErrInvalidAuthCommand
	}
	if err := s.ensureCredentialManager(ctx, query.ManagerID); err != nil {
		return CredentialManagementResult{}, err
	}
	policy, err := s.repo.FindStoreCredentialPolicy(ctx, query.StoreID)
	if err != nil {
		return CredentialManagementResult{}, err
	}
	actors, err := s.repo.ListActors(ctx)
	if err != nil {
		return CredentialManagementResult{}, err
	}
	return CredentialManagementResult{
		StoreID:     query.StoreID,
		StorePolicy: credentialPolicyResult(policy),
		Actors:      actorCredentialResults(actors),
	}, nil
}

func (s *CredentialManagementService) SetStoreCredentialPolicy(ctx context.Context, command SetStoreCredentialPolicyCommand) (CredentialPolicyResult, error) {
	if command.IdempotencyKey == "" {
		return CredentialPolicyResult{}, ErrIdempotencyKeyRequired
	}
	if command.StoreID == "" || command.ManagerID == "" {
		return CredentialPolicyResult{}, ErrInvalidAuthCommand
	}
	if err := s.ensureCredentialManager(ctx, command.ManagerID); err != nil {
		return CredentialPolicyResult{}, err
	}
	policy, err := credentialPolicyFromInput(command.Required, command.AllowedKinds)
	if err != nil {
		return CredentialPolicyResult{}, err
	}
	result := credentialPolicyResult(policy)
	fingerprint := credentialPolicyFingerprint(command.ManagerID, result)
	if result, found, err := s.findCredentialPolicyIdempotency(ctx, setStoreCredentialPolicyOperation, command.IdempotencyKey, command.StoreID, fingerprint); err != nil || found {
		return result, err
	}
	claimed := false
	record := IdempotencyRecord{
		Operation:   setStoreCredentialPolicyOperation,
		Key:         command.IdempotencyKey,
		TargetID:    command.StoreID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   time.Now().UTC(),
	}
	if err := RunTransaction(ctx, s.transactions, func(ctx context.Context) error {
		var err error
		claimed, err = s.idempotency.Claim(ctx, record)
		if err != nil || !claimed {
			return err
		}
		if err := s.repo.SaveStoreCredentialPolicy(ctx, command.StoreID, policy); err != nil {
			return err
		}
		return s.idempotency.Save(ctx, record)
	}); err != nil {
		return CredentialPolicyResult{}, err
	}
	if !claimed {
		if result, found, err := s.findCredentialPolicyIdempotency(ctx, setStoreCredentialPolicyOperation, command.IdempotencyKey, command.StoreID, fingerprint); err != nil || found {
			return result, err
		}
		return CredentialPolicyResult{}, ErrIdempotencyResultMissing
	}
	return result, nil
}

func (s *CredentialManagementService) SetActorCredentialPolicy(ctx context.Context, command SetActorCredentialPolicyCommand) (ActorCredentialResult, error) {
	if command.IdempotencyKey == "" {
		return ActorCredentialResult{}, ErrIdempotencyKeyRequired
	}
	if command.TargetActorID == "" || command.ManagerID == "" {
		return ActorCredentialResult{}, ErrInvalidAuthCommand
	}
	if command.TargetActorID == command.ManagerID {
		return ActorCredentialResult{}, ErrSeparationOfDutiesViolation
	}
	if err := s.ensureCredentialManager(ctx, command.ManagerID); err != nil {
		return ActorCredentialResult{}, err
	}
	actor, err := s.repo.FindActor(ctx, command.TargetActorID)
	if err != nil {
		return ActorCredentialResult{}, err
	}
	var policy *domain.CredentialPolicy
	if command.InheritStorePolicy {
		policy = nil
	} else {
		result, err := credentialPolicyFromInput(command.Required, command.AllowedKinds)
		if err != nil {
			return ActorCredentialResult{}, err
		}
		policy = &result
	}
	actor.CredentialPolicy = cloneCredentialPolicyPointer(policy)
	result := actorCredentialResult(actor)
	fingerprint := actorCredentialPolicyFingerprint(command.ManagerID, command.InheritStorePolicy, result)
	if result, found, err := s.findActorCredentialIdempotency(ctx, setActorCredentialPolicyOperation, command.IdempotencyKey, command.TargetActorID, fingerprint); err != nil || found {
		return result, err
	}
	claimed := false
	record := IdempotencyRecord{
		Operation:   setActorCredentialPolicyOperation,
		Key:         command.IdempotencyKey,
		TargetID:    command.TargetActorID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   time.Now().UTC(),
	}
	if err := RunTransaction(ctx, s.transactions, func(ctx context.Context) error {
		var err error
		claimed, err = s.idempotency.Claim(ctx, record)
		if err != nil || !claimed {
			return err
		}
		updatedActor, err := s.repo.UpdateActorCredentialPolicy(ctx, command.TargetActorID, policy)
		if err != nil {
			return err
		}
		result = actorCredentialResult(updatedActor)
		record.Result = result
		return s.idempotency.Save(ctx, record)
	}); err != nil {
		return ActorCredentialResult{}, err
	}
	if !claimed {
		if result, found, err := s.findActorCredentialIdempotency(ctx, setActorCredentialPolicyOperation, command.IdempotencyKey, command.TargetActorID, fingerprint); err != nil || found {
			return result, err
		}
		return ActorCredentialResult{}, ErrIdempotencyResultMissing
	}
	return result, nil
}

func (s *CredentialManagementService) AddCredentialBinding(ctx context.Context, command AddCredentialBindingCommand) (ActorCredentialResult, error) {
	if command.IdempotencyKey == "" {
		return ActorCredentialResult{}, ErrIdempotencyKeyRequired
	}
	token := strings.TrimSpace(command.Token)
	if command.TargetActorID == "" || command.ManagerID == "" || command.Kind == "" || token == "" {
		return ActorCredentialResult{}, ErrInvalidAuthCommand
	}
	if command.TargetActorID == command.ManagerID {
		return ActorCredentialResult{}, ErrSeparationOfDutiesViolation
	}
	if !isValidCredentialKind(command.Kind) {
		return ActorCredentialResult{}, ErrInvalidAuthCommand
	}
	if err := s.ensureCredentialManager(ctx, command.ManagerID); err != nil {
		return ActorCredentialResult{}, err
	}
	actor, err := s.repo.FindActor(ctx, command.TargetActorID)
	if err != nil {
		return ActorCredentialResult{}, err
	}
	tokenHash := hashCredentialToken(token)
	maskedToken := strings.TrimSpace(command.MaskedToken)
	if maskedToken == "" {
		maskedToken = defaultMaskedToken(command.Kind, tokenHash)
	}
	actor.CredentialBindings = addCredentialBinding(actor.CredentialBindings, command.Kind, tokenHash, maskedToken)
	result := actorCredentialResult(actor)
	fingerprint := credentialBindingFingerprint(command.ManagerID, command.Kind, tokenFingerprint(tokenHash), maskedToken)
	if result, found, err := s.findActorCredentialIdempotency(ctx, addCredentialBindingOperation, command.IdempotencyKey, command.TargetActorID, fingerprint); err != nil || found {
		return result, err
	}
	claimed := false
	record := IdempotencyRecord{
		Operation:   addCredentialBindingOperation,
		Key:         command.IdempotencyKey,
		TargetID:    command.TargetActorID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   time.Now().UTC(),
	}
	if err := RunTransaction(ctx, s.transactions, func(ctx context.Context) error {
		var err error
		claimed, err = s.idempotency.Claim(ctx, record)
		if err != nil || !claimed {
			return err
		}
		updatedActor, err := s.repo.UpdateActorCredentialBindings(ctx, command.TargetActorID, func(bindings []domain.CredentialBinding) ([]domain.CredentialBinding, error) {
			return addCredentialBinding(bindings, command.Kind, tokenHash, maskedToken), nil
		})
		if err != nil {
			return err
		}
		result = actorCredentialResult(updatedActor)
		record.Result = result
		return s.idempotency.Save(ctx, record)
	}); err != nil {
		return ActorCredentialResult{}, err
	}
	if !claimed {
		if result, found, err := s.findActorCredentialIdempotency(ctx, addCredentialBindingOperation, command.IdempotencyKey, command.TargetActorID, fingerprint); err != nil || found {
			return result, err
		}
		return ActorCredentialResult{}, ErrIdempotencyResultMissing
	}
	return result, nil
}

func (s *CredentialManagementService) RevokeCredentialBinding(ctx context.Context, command RevokeCredentialBindingCommand) (ActorCredentialResult, error) {
	if command.IdempotencyKey == "" {
		return ActorCredentialResult{}, ErrIdempotencyKeyRequired
	}
	if command.TargetActorID == "" || command.ManagerID == "" || command.Kind == "" || command.TokenFingerprint == "" {
		return ActorCredentialResult{}, ErrInvalidAuthCommand
	}
	if command.TargetActorID == command.ManagerID {
		return ActorCredentialResult{}, ErrSeparationOfDutiesViolation
	}
	if !isValidCredentialKind(command.Kind) {
		return ActorCredentialResult{}, ErrInvalidAuthCommand
	}
	if err := s.ensureCredentialManager(ctx, command.ManagerID); err != nil {
		return ActorCredentialResult{}, err
	}
	actor, err := s.repo.FindActor(ctx, command.TargetActorID)
	if err != nil {
		return ActorCredentialResult{}, err
	}
	updatedBindings, err := revokeCredentialBinding(actor.CredentialBindings, command.Kind, command.TokenFingerprint)
	if err != nil {
		return ActorCredentialResult{}, err
	}
	actor.CredentialBindings = updatedBindings
	result := actorCredentialResult(actor)
	fingerprint := credentialBindingFingerprint(command.ManagerID, command.Kind, command.TokenFingerprint, "")
	if result, found, err := s.findActorCredentialIdempotency(ctx, revokeCredentialBindingOperation, command.IdempotencyKey, command.TargetActorID, fingerprint); err != nil || found {
		return result, err
	}
	claimed := false
	record := IdempotencyRecord{
		Operation:   revokeCredentialBindingOperation,
		Key:         command.IdempotencyKey,
		TargetID:    command.TargetActorID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   time.Now().UTC(),
	}
	if err := RunTransaction(ctx, s.transactions, func(ctx context.Context) error {
		var err error
		claimed, err = s.idempotency.Claim(ctx, record)
		if err != nil || !claimed {
			return err
		}
		updatedActor, err := s.repo.UpdateActorCredentialBindings(ctx, command.TargetActorID, func(bindings []domain.CredentialBinding) ([]domain.CredentialBinding, error) {
			return revokeCredentialBinding(bindings, command.Kind, command.TokenFingerprint)
		})
		if err != nil {
			return err
		}
		result = actorCredentialResult(updatedActor)
		record.Result = result
		return s.idempotency.Save(ctx, record)
	}); err != nil {
		return ActorCredentialResult{}, err
	}
	if !claimed {
		if result, found, err := s.findActorCredentialIdempotency(ctx, revokeCredentialBindingOperation, command.IdempotencyKey, command.TargetActorID, fingerprint); err != nil || found {
			return result, err
		}
		return ActorCredentialResult{}, ErrIdempotencyResultMissing
	}
	return result, nil
}

func (s *CredentialManagementService) findCredentialPolicyIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string) (CredentialPolicyResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, operation, key)
	if err != nil || !found {
		return CredentialPolicyResult{}, found, err
	}
	if record.TargetID != targetID || record.Fingerprint != fingerprint {
		return CredentialPolicyResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(CredentialPolicyResult)
	if !ok {
		return CredentialPolicyResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

func (s *CredentialManagementService) findActorCredentialIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string) (ActorCredentialResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, operation, key)
	if err != nil || !found {
		return ActorCredentialResult{}, found, err
	}
	if record.TargetID != targetID || record.Fingerprint != fingerprint {
		return ActorCredentialResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(ActorCredentialResult)
	if !ok {
		return ActorCredentialResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

func (s *CredentialManagementService) ensureCredentialManager(ctx context.Context, managerID string) error {
	manager, err := s.repo.FindActor(ctx, managerID)
	if err != nil {
		return err
	}
	return CheckPermission(manager.Roles, PermissionCredentialsManage)
}

func credentialPolicyFromInput(required bool, allowedKinds []domain.CredentialKind) (domain.CredentialPolicy, error) {
	uniqueKinds := make([]domain.CredentialKind, 0, len(allowedKinds))
	seen := map[domain.CredentialKind]struct{}{}
	for _, kind := range allowedKinds {
		if !isValidCredentialKind(kind) {
			return domain.CredentialPolicy{}, ErrInvalidAuthCommand
		}
		if _, ok := seen[kind]; ok {
			continue
		}
		seen[kind] = struct{}{}
		uniqueKinds = append(uniqueKinds, kind)
	}
	if required && len(uniqueKinds) == 0 {
		return domain.CredentialPolicy{}, ErrInvalidAuthCommand
	}
	return domain.CredentialPolicy{Required: required, AllowedKinds: uniqueKinds}, nil
}

func cloneCredentialPolicyPointer(policy *domain.CredentialPolicy) *domain.CredentialPolicy {
	if policy == nil {
		return nil
	}
	clone := domain.CredentialPolicy{
		Required:     policy.Required,
		AllowedKinds: append([]domain.CredentialKind(nil), policy.AllowedKinds...),
	}
	return &clone
}

func addCredentialBinding(bindings []domain.CredentialBinding, kind domain.CredentialKind, tokenHash string, maskedToken string) []domain.CredentialBinding {
	updatedBindings := append([]domain.CredentialBinding(nil), bindings...)
	updated := false
	for i := range updatedBindings {
		binding := &updatedBindings[i]
		if binding.Kind == kind && binding.TokenHash == tokenHash {
			binding.MaskedToken = maskedToken
			binding.Active = true
			updated = true
			break
		}
	}
	if !updated {
		updatedBindings = append(updatedBindings, domain.CredentialBinding{
			Kind:        kind,
			TokenHash:   tokenHash,
			MaskedToken: maskedToken,
			Active:      true,
		})
	}
	return updatedBindings
}

func revokeCredentialBinding(bindings []domain.CredentialBinding, kind domain.CredentialKind, fingerprint string) ([]domain.CredentialBinding, error) {
	updatedBindings := append([]domain.CredentialBinding(nil), bindings...)
	found := false
	for i := range updatedBindings {
		binding := &updatedBindings[i]
		if binding.Kind == kind && tokenFingerprint(binding.TokenHash) == fingerprint {
			binding.Active = false
			found = true
			break
		}
	}
	if !found {
		return nil, ErrCredentialBindingNotFound
	}
	return updatedBindings, nil
}

func credentialPolicyFingerprint(managerID string, policy CredentialPolicyResult) string {
	return fmt.Sprintf("%s|%t|%s", managerID, policy.Required, credentialKindsFingerprint(policy.AllowedKinds))
}

func actorCredentialPolicyFingerprint(managerID string, inheritStorePolicy bool, actor ActorCredentialResult) string {
	if actor.CredentialPolicy == nil {
		return fmt.Sprintf("%s|%s|%t", managerID, actor.ID, inheritStorePolicy)
	}
	return fmt.Sprintf("%s|%s|%t|%s", managerID, actor.ID, inheritStorePolicy, credentialPolicyFingerprint("", *actor.CredentialPolicy))
}

func credentialBindingFingerprint(managerID string, kind domain.CredentialKind, tokenFingerprint string, maskedToken string) string {
	return fmt.Sprintf("%s|%s|%s|%s", managerID, kind, tokenFingerprint, maskedToken)
}

func credentialKindsFingerprint(kinds []domain.CredentialKind) string {
	parts := make([]string, 0, len(kinds))
	for _, kind := range kinds {
		parts = append(parts, string(kind))
	}
	return strings.Join(parts, ",")
}

func isValidCredentialKind(kind domain.CredentialKind) bool {
	switch kind {
	case domain.CredentialKindIButton, domain.CredentialKindMSRCard, domain.CredentialKindBarcodeCard:
		return true
	default:
		return false
	}
}

func actorCredentialResults(actors []domain.Actor) []ActorCredentialResult {
	results := make([]ActorCredentialResult, 0, len(actors))
	for _, actor := range actors {
		results = append(results, actorCredentialResult(actor))
	}
	return results
}

func actorCredentialResult(actor domain.Actor) ActorCredentialResult {
	var policy *CredentialPolicyResult
	if actor.CredentialPolicy != nil {
		result := credentialPolicyResult(*actor.CredentialPolicy)
		policy = &result
	}
	return ActorCredentialResult{
		ID:                 actor.ID,
		Roles:              append([]domain.Role(nil), actor.Roles...),
		CredentialPolicy:   policy,
		CredentialBindings: credentialBindingResults(actor.CredentialBindings),
	}
}

func credentialPolicyResult(policy domain.CredentialPolicy) CredentialPolicyResult {
	return CredentialPolicyResult{
		Required:     policy.Required,
		AllowedKinds: append([]domain.CredentialKind(nil), policy.AllowedKinds...),
	}
}

func credentialBindingResults(bindings []domain.CredentialBinding) []CredentialBindingResult {
	results := make([]CredentialBindingResult, 0, len(bindings))
	for _, binding := range bindings {
		results = append(results, CredentialBindingResult{
			Kind:             binding.Kind,
			TokenFingerprint: tokenFingerprint(binding.TokenHash),
			MaskedToken:      binding.MaskedToken,
			Active:           binding.Active,
		})
	}
	return results
}

func defaultMaskedToken(kind domain.CredentialKind, tokenHash string) string {
	return string(kind) + " demo ****" + tokenFingerprint(tokenHash)
}
