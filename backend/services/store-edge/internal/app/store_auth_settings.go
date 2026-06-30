package app

import (
	"context"
	"fmt"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

const setStoreAuthSettingsOperation = "store_settings.set_auth_settings"

type StoreAuthSettingsRepository interface {
	FindStoreAuthSettings(ctx context.Context, storeID string) (domain.StoreAuthSettings, error)
	SaveStoreAuthSettings(ctx context.Context, settings domain.StoreAuthSettings) error
	FindActor(ctx context.Context, actorID string) (domain.Actor, error)
}

type StoreAuthSettingsService struct {
	repo         StoreAuthSettingsRepository
	idempotency  IdempotencyStore
	transactions TransactionRunner
	now          func() time.Time
}

type StoreAuthSettingsOption func(*StoreAuthSettingsService)

func NewStoreAuthSettingsService(repo StoreAuthSettingsRepository, idempotency IdempotencyStore, options ...StoreAuthSettingsOption) *StoreAuthSettingsService {
	service := &StoreAuthSettingsService{
		repo:        repo,
		idempotency: idempotency,
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
