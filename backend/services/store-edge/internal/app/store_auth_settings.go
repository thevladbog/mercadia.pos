package app

import (
	"context"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

type StoreAuthSettingsRepository interface {
	FindStoreAuthSettings(ctx context.Context, storeID string) (domain.StoreAuthSettings, error)
	SaveStoreAuthSettings(ctx context.Context, settings domain.StoreAuthSettings) error
	FindActor(ctx context.Context, actorID string) (domain.Actor, error)
}

type StoreAuthSettingsService struct {
	repo StoreAuthSettingsRepository
	now  func() time.Time
}

func NewStoreAuthSettingsService(repo StoreAuthSettingsRepository) *StoreAuthSettingsService {
	return &StoreAuthSettingsService{
		repo: repo,
		now: func() time.Time {
			return time.Now().UTC()
		},
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
	if err := s.repo.SaveStoreAuthSettings(ctx, settings); err != nil {
		return StoreAuthSettingsResult{}, err
	}
	return storeAuthSettingsResult(settings), nil
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
