package app

import (
	"context"
	"fmt"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

const (
	setStoreAuthSettingsOperation = "store_settings.set_auth_settings"
	resetAuthLockoutOperation     = "store_settings.reset_auth_lockout"
	AuthFailureReasonLockoutReset = "lockout_reset"
)

type StoreAuthSettingsRepository interface {
	FindStoreAuthSettings(ctx context.Context, storeID string) (domain.StoreAuthSettings, error)
	SaveStoreAuthSettings(ctx context.Context, settings domain.StoreAuthSettings) error
	ListAuthAttempts(ctx context.Context, filter AuthAttemptFilter, params PageParams) (PageResult[domain.AuthAttempt], error)
	SaveAuthAttempt(ctx context.Context, attempt domain.AuthAttempt) error
	FindActor(ctx context.Context, actorID string) (domain.Actor, error)
}

type StoreAuthSettingsService struct {
	repo         StoreAuthSettingsRepository
	idempotency  IdempotencyStore
	transactions TransactionRunner
	now          func() time.Time
	newAttemptID func(prefix string) string
}

type StoreAuthSettingsOption func(*StoreAuthSettingsService)

func NewStoreAuthSettingsService(repo StoreAuthSettingsRepository, idempotency IdempotencyStore, options ...StoreAuthSettingsOption) *StoreAuthSettingsService {
	service := &StoreAuthSettingsService{
		repo:         repo,
		idempotency:  idempotency,
		newAttemptID: randomID,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithStoreAuthSettingsTransactionRunner(runner TransactionRunner) StoreAuthSettingsOption {
	return func(service *StoreAuthSettingsService) {
		service.transactions = runner
	}
}

type StoreAuthSettingsResult struct {
	StoreID                string
	FailedAttemptLimit     int
	LockoutDurationSeconds int
	POSAutoLockSeconds     int
	UpdatedByID            string
	UpdatedAt              time.Time
}

type SetStoreAuthSettingsCommand struct {
	IdempotencyKey         string
	StoreID                string
	ManagerID              string
	FailedAttemptLimit     int
	LockoutDurationSeconds int
	POSAutoLockSeconds     int
}

type AuthAttemptFilter struct {
	StoreID    string
	ActorID    string
	TerminalID string
	Successful *bool
	Since      time.Time
	Until      time.Time
}

type ListAuthAttemptsQuery struct {
	StoreID    string
	ManagerID  string
	ActorID    string
	TerminalID string
	Successful *bool
	Since      time.Time
	Until      time.Time
	Page       PageParams
}

type AuthAttemptResult struct {
	ID                    string
	StoreID               string
	ActorID               string
	TerminalID            string
	CredentialKind        domain.CredentialKind
	CredentialFingerprint string
	Successful            bool
	FailureReason         string
	CreatedAt             time.Time
}

type ResetAuthLockoutCommand struct {
	IdempotencyKey string
	StoreID        string
	ActorID        string
	ManagerID      string
	Reason         string
}

type AuthLockoutResetResult struct {
	StoreID   string
	ActorID   string
	ResetByID string
	Reason    string
	ResetAt   time.Time
}

func (s *StoreAuthSettingsService) GetStoreAuthSettings(ctx context.Context, storeID string) (StoreAuthSettingsResult, error) {
	if storeID == "" {
		return StoreAuthSettingsResult{}, ErrInvalidAuthCommand
	}
	settings, err := s.repo.FindStoreAuthSettings(ctx, storeID)
	if err != nil {
		return StoreAuthSettingsResult{}, err
	}
	return storeAuthSettingsResult(settings), nil
}

func (s *StoreAuthSettingsService) ListAuthAttempts(ctx context.Context, query ListAuthAttemptsQuery) (PageResult[AuthAttemptResult], error) {
	if query.StoreID == "" || query.ManagerID == "" {
		return PageResult[AuthAttemptResult]{}, ErrInvalidAuthCommand
	}
	manager, err := s.repo.FindActor(ctx, query.ManagerID)
	if err != nil {
		return PageResult[AuthAttemptResult]{}, err
	}
	if err := CheckPermission(manager.Roles, PermissionStoreSettingsManage); err != nil {
		return PageResult[AuthAttemptResult]{}, err
	}
	attempts, err := s.repo.ListAuthAttempts(ctx, AuthAttemptFilter{
		StoreID:    query.StoreID,
		ActorID:    query.ActorID,
		TerminalID: query.TerminalID,
		Successful: query.Successful,
		Since:      query.Since,
		Until:      query.Until,
	}, query.Page)
	if err != nil {
		return PageResult[AuthAttemptResult]{}, err
	}
	results := make([]AuthAttemptResult, 0, len(attempts.Items))
	for _, attempt := range attempts.Items {
		results = append(results, authAttemptResult(attempt))
	}
	return PageResult[AuthAttemptResult]{Items: results, TotalCount: attempts.TotalCount}, nil
}

func (s *StoreAuthSettingsService) SetStoreAuthSettings(ctx context.Context, command SetStoreAuthSettingsCommand) (StoreAuthSettingsResult, error) {
	if command.IdempotencyKey == "" {
		return StoreAuthSettingsResult{}, ErrIdempotencyKeyRequired
	}
	if command.StoreID == "" || command.ManagerID == "" {
		return StoreAuthSettingsResult{}, ErrInvalidAuthCommand
	}
	manager, err := s.repo.FindActor(ctx, command.ManagerID)
	if err != nil {
		return StoreAuthSettingsResult{}, err
	}
	if err := CheckPermission(manager.Roles, PermissionStoreSettingsManage); err != nil {
		return StoreAuthSettingsResult{}, err
	}
	fingerprint := fmt.Sprintf("%s|%s|%d|%d|%d", command.StoreID, command.ManagerID,
		command.FailedAttemptLimit, command.LockoutDurationSeconds, command.POSAutoLockSeconds)
	if result, found, err := s.findStoreAuthSettingsIdempotency(ctx, command.IdempotencyKey, command.StoreID, fingerprint); err != nil || found {
		return result, err
	}
	settings, err := domain.NewStoreAuthSettings(domain.CreateStoreAuthSettingsInput{
		StoreID:                command.StoreID,
		FailedAttemptLimit:     command.FailedAttemptLimit,
		LockoutDurationSeconds: command.LockoutDurationSeconds,
		POSAutoLockSeconds:     command.POSAutoLockSeconds,
		UpdatedByID:            command.ManagerID,
		Now:                    s.now(),
	})
	if err != nil {
		return StoreAuthSettingsResult{}, err
	}
	result := storeAuthSettingsResult(settings)
	if err := RunTransaction(ctx, s.transactions, func(ctx context.Context) error {
		if err := s.repo.SaveStoreAuthSettings(ctx, settings); err != nil {
			return err
		}
		return s.idempotency.Save(ctx, IdempotencyRecord{
			Operation:   setStoreAuthSettingsOperation,
			Key:         command.IdempotencyKey,
			TargetID:    command.StoreID,
			Fingerprint: fingerprint,
			Result:      result,
			CreatedAt:   s.now(),
		})
	}); err != nil {
		return StoreAuthSettingsResult{}, err
	}
	return result, nil
}

func (s *StoreAuthSettingsService) ResetAuthLockout(ctx context.Context, command ResetAuthLockoutCommand) (AuthLockoutResetResult, error) {
	if command.IdempotencyKey == "" {
		return AuthLockoutResetResult{}, ErrIdempotencyKeyRequired
	}
	if command.StoreID == "" || command.ActorID == "" || command.ManagerID == "" {
		return AuthLockoutResetResult{}, ErrInvalidAuthCommand
	}
	manager, err := s.repo.FindActor(ctx, command.ManagerID)
	if err != nil {
		return AuthLockoutResetResult{}, err
	}
	if err := CheckPermission(manager.Roles, PermissionStoreSettingsManage); err != nil {
		return AuthLockoutResetResult{}, err
	}
	if _, err := s.repo.FindActor(ctx, command.ActorID); err != nil {
		return AuthLockoutResetResult{}, err
	}
	fingerprint := fmt.Sprintf("%s|%s|%s|%s", command.StoreID, command.ActorID, command.ManagerID, command.Reason)
	targetID := command.StoreID + "|" + command.ActorID
	if result, found, err := s.findAuthLockoutResetIdempotency(ctx, command.IdempotencyKey, targetID, fingerprint); err != nil || found {
		return result, err
	}
	now := s.now()
	attempt, err := domain.NewAuthAttempt(domain.CreateAuthAttemptInput{
		ID:            s.newAttemptID("auth_attempt"),
		StoreID:       command.StoreID,
		ActorID:       command.ActorID,
		Successful:    true,
		FailureReason: AuthFailureReasonLockoutReset,
		Now:           now,
	})
	if err != nil {
		return AuthLockoutResetResult{}, err
	}
	result := AuthLockoutResetResult{
		StoreID:   command.StoreID,
		ActorID:   command.ActorID,
		ResetByID: command.ManagerID,
		Reason:    command.Reason,
		ResetAt:   now,
	}
	claimed := false
	record := IdempotencyRecord{
		Operation:   resetAuthLockoutOperation,
		Key:         command.IdempotencyKey,
		TargetID:    targetID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   now,
	}
	if err := RunTransaction(ctx, s.transactions, func(ctx context.Context) error {
		var err error
		claimed, err = s.idempotency.Claim(ctx, record)
		if err != nil || !claimed {
			return err
		}
		if err := s.repo.SaveAuthAttempt(ctx, attempt); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return AuthLockoutResetResult{}, err
	}
	if !claimed {
		if result, found, err := s.findAuthLockoutResetIdempotency(ctx, command.IdempotencyKey, targetID, fingerprint); err != nil || found {
			return result, err
		}
		return AuthLockoutResetResult{}, ErrIdempotencyResultMissing
	}
	return result, nil
}

func (s *StoreAuthSettingsService) findStoreAuthSettingsIdempotency(ctx context.Context, key string, targetID string, fingerprint string) (StoreAuthSettingsResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, setStoreAuthSettingsOperation, key)
	if err != nil || !found {
		return StoreAuthSettingsResult{}, found, err
	}
	if record.TargetID != targetID || record.Fingerprint != fingerprint {
		return StoreAuthSettingsResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(StoreAuthSettingsResult)
	if !ok {
		return StoreAuthSettingsResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

func (s *StoreAuthSettingsService) findAuthLockoutResetIdempotency(ctx context.Context, key string, targetID string, fingerprint string) (AuthLockoutResetResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, resetAuthLockoutOperation, key)
	if err != nil || !found {
		return AuthLockoutResetResult{}, found, err
	}
	if record.TargetID != targetID || record.Fingerprint != fingerprint {
		return AuthLockoutResetResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(AuthLockoutResetResult)
	if !ok {
		return AuthLockoutResetResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

func authAttemptResult(attempt domain.AuthAttempt) AuthAttemptResult {
	return AuthAttemptResult{
		ID:                    attempt.ID,
		StoreID:               attempt.StoreID,
		ActorID:               attempt.ActorID,
		TerminalID:            attempt.TerminalID,
		CredentialKind:        attempt.CredentialKind,
		CredentialFingerprint: attempt.CredentialFingerprint,
		Successful:            attempt.Successful,
		FailureReason:         attempt.FailureReason,
		CreatedAt:             attempt.CreatedAt,
	}
}

func storeAuthSettingsResult(settings domain.StoreAuthSettings) StoreAuthSettingsResult {
	return StoreAuthSettingsResult{
		StoreID:                settings.StoreID,
		FailedAttemptLimit:     settings.FailedAttemptLimit,
		LockoutDurationSeconds: settings.LockoutDurationSeconds,
		POSAutoLockSeconds:     settings.POSAutoLockSeconds,
		UpdatedByID:            settings.UpdatedByID,
		UpdatedAt:              settings.UpdatedAt,
	}
}
