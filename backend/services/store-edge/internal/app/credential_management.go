package app

import (
	"context"
	"errors"
	"strings"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

var ErrCredentialBindingNotFound = errors.New("credential binding not found")

type CredentialManagementRepository interface {
	FindActor(ctx context.Context, actorID string) (domain.Actor, error)
	ListActors(ctx context.Context) ([]domain.Actor, error)
	SaveActor(ctx context.Context, actor domain.Actor) error
	FindStoreCredentialPolicy(ctx context.Context, storeID string) (domain.CredentialPolicy, error)
	SaveStoreCredentialPolicy(ctx context.Context, storeID string, policy domain.CredentialPolicy) error
}

type CredentialManagementService struct {
	repo CredentialManagementRepository
}

func NewCredentialManagementService(repo CredentialManagementRepository) *CredentialManagementService {
	return &CredentialManagementService{repo: repo}
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
	StoreID      string
	ManagerID    string
	Required     bool
	AllowedKinds []domain.CredentialKind
}

type SetActorCredentialPolicyCommand struct {
	TargetActorID      string
	ManagerID          string
	InheritStorePolicy bool
	Required           bool
	AllowedKinds       []domain.CredentialKind
}

type AddCredentialBindingCommand struct {
	TargetActorID string
	ManagerID     string
	Kind          domain.CredentialKind
	Token         string
	MaskedToken   string
}

type RevokeCredentialBindingCommand struct {
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
	if err := s.repo.SaveStoreCredentialPolicy(ctx, command.StoreID, policy); err != nil {
		return CredentialPolicyResult{}, err
	}
	return credentialPolicyResult(policy), nil
}

func (s *CredentialManagementService) SetActorCredentialPolicy(ctx context.Context, command SetActorCredentialPolicyCommand) (ActorCredentialResult, error) {
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
	if command.InheritStorePolicy {
		actor.CredentialPolicy = nil
	} else {
		policy, err := credentialPolicyFromInput(command.Required, command.AllowedKinds)
		if err != nil {
			return ActorCredentialResult{}, err
		}
		actor.CredentialPolicy = &policy
	}
	if err := s.repo.SaveActor(ctx, actor); err != nil {
		return ActorCredentialResult{}, err
	}
	return actorCredentialResult(actor), nil
}

func (s *CredentialManagementService) AddCredentialBinding(ctx context.Context, command AddCredentialBindingCommand) (ActorCredentialResult, error) {
	if command.TargetActorID == "" || command.ManagerID == "" || command.Kind == "" || strings.TrimSpace(command.Token) == "" {
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
	tokenHash := hashCredentialToken(command.Token)
	maskedToken := strings.TrimSpace(command.MaskedToken)
	if maskedToken == "" {
		maskedToken = defaultMaskedToken(command.Kind, tokenHash)
	}
	updated := false
	for i := range actor.CredentialBindings {
		binding := &actor.CredentialBindings[i]
		if binding.Kind == command.Kind && binding.TokenHash == tokenHash {
			binding.MaskedToken = maskedToken
			binding.Active = true
			updated = true
			break
		}
	}
	if !updated {
		actor.CredentialBindings = append(actor.CredentialBindings, domain.CredentialBinding{
			Kind:        command.Kind,
			TokenHash:   tokenHash,
			MaskedToken: maskedToken,
			Active:      true,
		})
	}
	if err := s.repo.SaveActor(ctx, actor); err != nil {
		return ActorCredentialResult{}, err
	}
	return actorCredentialResult(actor), nil
}

func (s *CredentialManagementService) RevokeCredentialBinding(ctx context.Context, command RevokeCredentialBindingCommand) (ActorCredentialResult, error) {
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
	found := false
	for i := range actor.CredentialBindings {
		binding := &actor.CredentialBindings[i]
		if binding.Kind == command.Kind && tokenFingerprint(binding.TokenHash) == command.TokenFingerprint {
			binding.Active = false
			found = true
			break
		}
	}
	if !found {
		return ActorCredentialResult{}, ErrCredentialBindingNotFound
	}
	if err := s.repo.SaveActor(ctx, actor); err != nil {
		return ActorCredentialResult{}, err
	}
	return actorCredentialResult(actor), nil
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
