package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/domain"
)

var (
	ErrStoreNotFound           = errors.New("store not found")
	ErrInvalidStoreCommand     = errors.New("invalid store command")
	ErrInvalidStoreRegistryCmd = errors.New("invalid store registry command")
)

type StoreRepository interface {
	SaveStore(ctx context.Context, store domain.Store) error
	FindStore(ctx context.Context, storeID string) (domain.Store, error)
	ListStores(ctx context.Context) ([]domain.Store, error)
	CountStores(ctx context.Context) (int, error)
}

type StoreRegistryService struct {
	stores      StoreRepository
	idempotency IdempotencyStore
	now         func() time.Time
	newID       func(prefix string) string
}

func NewStoreRegistryService(stores StoreRepository, idempotency IdempotencyStore) *StoreRegistryService {
	return &StoreRegistryService{
		stores:      stores,
		idempotency: idempotency,
		now:         time.Now,
		newID:       defaultNewID,
	}
}

type RegisterStoreCommand struct {
	StoreID        string
	Name           string
	Region         string
	IdempotencyKey string
}

type StoreResult struct {
	Store domain.Store
}

func (s *StoreRegistryService) RegisterStore(ctx context.Context, command RegisterStoreCommand) (StoreResult, error) {
	const operation = "register_store"
	if command.StoreID == "" || command.Name == "" {
		return StoreResult{}, ErrInvalidStoreRegistryCmd
	}
	if command.IdempotencyKey == "" {
		return StoreResult{}, ErrIdempotencyKeyRequired
	}

	fingerprint := registerStoreFingerprint(command)
	if result, found, err := s.findStoreIdempotency(ctx, operation, command.IdempotencyKey, command.StoreID, fingerprint); err != nil || found {
		return result, err
	}

	existing, err := s.stores.FindStore(ctx, command.StoreID)
	if err == nil {
		result := StoreResult{Store: existing}
		if saveErr := s.saveStoreIdempotency(ctx, operation, command.IdempotencyKey, command.StoreID, fingerprint, result); saveErr != nil {
			return StoreResult{}, saveErr
		}
		return result, nil
	}
	if !errors.Is(err, ErrStoreNotFound) {
		return StoreResult{}, err
	}

	now := s.now().UTC()
	store, err := domain.NewStore(domain.Store{
		ID:           command.StoreID,
		Name:         command.Name,
		Region:       command.Region,
		RegisteredAt: now,
		UpdatedAt:    now,
	})
	if err != nil {
		return StoreResult{}, ErrInvalidStoreCommand
	}
	if err := s.stores.SaveStore(ctx, store); err != nil {
		return StoreResult{}, err
	}

	result := StoreResult{Store: store}
	if err := s.saveStoreIdempotency(ctx, operation, command.IdempotencyKey, command.StoreID, fingerprint, result); err != nil {
		return StoreResult{}, err
	}
	return result, nil
}

func (s *StoreRegistryService) ListStores(ctx context.Context) ([]domain.Store, error) {
	return s.stores.ListStores(ctx)
}

func (s *StoreRegistryService) CountStores(ctx context.Context) (int, error) {
	return s.stores.CountStores(ctx)
}

func (s *StoreRegistryService) findStoreIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string) (StoreResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, operation, key)
	if err != nil || !found {
		return StoreResult{}, found, err
	}
	if record.Fingerprint != fingerprint {
		return StoreResult{}, true, ErrIdempotencyKeyReused
	}
	if record.TargetID != "" && targetID != "" && record.TargetID != targetID {
		return StoreResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(StoreResult)
	if !ok {
		return StoreResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

func (s *StoreRegistryService) saveStoreIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string, result StoreResult) error {
	return s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         key,
		TargetID:    targetID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now().UTC(),
	})
}

func registerStoreFingerprint(command RegisterStoreCommand) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s", command.StoreID, command.Name, command.Region)))
	return hex.EncodeToString(sum[:])
}

func defaultNewID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UTC().UnixNano())
}
