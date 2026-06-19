package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/domain"
)

var (
	ErrInvalidSyncCommand = errors.New("invalid sync command")
	ErrSyncEventDuplicate = errors.New("sync event already accepted")
)

type SyncEventRepository interface {
	SaveSyncEvent(ctx context.Context, event domain.SyncEvent) error
	ExistsSyncEvent(ctx context.Context, storeID string, sourceEventID string) (bool, error)
}

type SyncEventInput struct {
	EventID    string
	EventType  string
	OccurredAt time.Time
	Payload    json.RawMessage
}

type AcceptSyncEventsCommand struct {
	StoreID        string
	IdempotencyKey string
	Events         []SyncEventInput
}

type SyncEventsResult struct {
	StoreID  string
	Status   string
	Accepted int
}

type SyncService struct {
	stores      StoreRepository
	syncEvents  SyncEventRepository
	catalog     CatalogProductRepository
	idempotency IdempotencyStore
	now         func() time.Time
	newID       func(prefix string) string
}

func NewSyncService(stores StoreRepository, syncEvents SyncEventRepository, catalog CatalogProductRepository, idempotency IdempotencyStore) *SyncService {
	return &SyncService{
		stores:      stores,
		syncEvents:  syncEvents,
		catalog:     catalog,
		idempotency: idempotency,
		now:         time.Now,
		newID:       defaultNewID,
	}
}

func (s *SyncService) AcceptEvents(ctx context.Context, command AcceptSyncEventsCommand) (SyncEventsResult, error) {
	const operation = "accept_sync_events"
	if command.StoreID == "" || len(command.Events) == 0 {
		return SyncEventsResult{}, ErrInvalidSyncCommand
	}
	if command.IdempotencyKey == "" {
		return SyncEventsResult{}, ErrIdempotencyKeyRequired
	}

	if _, err := s.stores.FindStore(ctx, command.StoreID); err != nil {
		return SyncEventsResult{}, err
	}

	fingerprint := syncEventsFingerprint(command)
	if result, found, err := s.findSyncIdempotency(ctx, operation, command.IdempotencyKey, command.StoreID, fingerprint); err != nil || found {
		return result, err
	}

	accepted := 0
	now := s.now().UTC()
	for _, input := range command.Events {
		if input.EventID == "" || input.EventType == "" {
			return SyncEventsResult{}, ErrInvalidSyncCommand
		}
		exists, err := s.syncEvents.ExistsSyncEvent(ctx, command.StoreID, input.EventID)
		if err != nil {
			return SyncEventsResult{}, err
		}
		if exists {
			continue
		}

		occurredAt := input.OccurredAt
		if occurredAt.IsZero() {
			occurredAt = now
		}

		event, err := domain.NewSyncEvent(domain.SyncEvent{
			ID:            s.newID("sync"),
			StoreID:       command.StoreID,
			EventType:     input.EventType,
			SourceEventID: input.EventID,
			Payload:       input.Payload,
			OccurredAt:    occurredAt.UTC(),
			ReceivedAt:    now,
		})
		if err != nil {
			return SyncEventsResult{}, ErrInvalidSyncCommand
		}
		if err := s.syncEvents.SaveSyncEvent(ctx, event); err != nil {
			if errors.Is(err, ErrSyncEventDuplicate) {
				continue
			}
			return SyncEventsResult{}, err
		}
		if err := s.applyCatalogEvent(ctx, event); err != nil {
			return SyncEventsResult{}, err
		}
		accepted++
	}

	result := SyncEventsResult{
		StoreID:  command.StoreID,
		Status:   "accepted",
		Accepted: accepted,
	}
	if err := s.saveSyncIdempotency(ctx, operation, command.IdempotencyKey, command.StoreID, fingerprint, result); err != nil {
		return SyncEventsResult{}, err
	}
	return result, nil
}

func (s *SyncService) applyCatalogEvent(ctx context.Context, event domain.SyncEvent) error {
	switch event.EventType {
	case "catalog.product.upserted":
		return s.upsertCatalogProductFromPayload(ctx, event)
	default:
		return nil
	}
}

func (s *SyncService) upsertCatalogProductFromPayload(ctx context.Context, event domain.SyncEvent) error {
	var payload struct {
		ProductID      string   `json:"productId"`
		Name           string   `json:"name"`
		Barcodes       []string `json:"barcodes"`
		UnitPriceMinor int64    `json:"unitPriceMinor"`
		TaxCategoryID  string   `json:"taxCategoryId"`
		Active         *bool    `json:"active"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return ErrInvalidSyncCommand
	}

	active := true
	if payload.Active != nil {
		active = *payload.Active
	}

	existing, err := s.catalog.FindProduct(ctx, event.StoreID, payload.ProductID)
	version := int64(1)
	if err == nil {
		version = existing.Version + 1
	} else if !errors.Is(err, ErrCatalogProductNotFound) {
		return err
	}

	product, err := domain.NewCatalogProduct(domain.CatalogProduct{
		ID:             payload.ProductID,
		StoreID:        event.StoreID,
		Name:           payload.Name,
		Barcodes:       payload.Barcodes,
		UnitPriceMinor: payload.UnitPriceMinor,
		TaxCategoryID:  payload.TaxCategoryID,
		Active:         active,
		Version:        version,
		UpdatedAt:      event.ReceivedAt,
	})
	if err != nil {
		return ErrInvalidSyncCommand
	}
	return s.catalog.SaveProduct(ctx, product)
}

func (s *SyncService) findSyncIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string) (SyncEventsResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, operation, key)
	if err != nil || !found {
		return SyncEventsResult{}, found, err
	}
	if record.Fingerprint != fingerprint {
		return SyncEventsResult{}, true, ErrIdempotencyKeyReused
	}
	if record.TargetID != "" && targetID != "" && record.TargetID != targetID {
		return SyncEventsResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(SyncEventsResult)
	if !ok {
		return SyncEventsResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

func (s *SyncService) saveSyncIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string, result SyncEventsResult) error {
	return s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         key,
		TargetID:    targetID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now().UTC(),
	})
}

func syncEventsFingerprint(command AcceptSyncEventsCommand) string {
	sum := sha256.New()
	_, _ = fmt.Fprintf(sum, "%s|", command.StoreID)
	for _, event := range command.Events {
		_, _ = fmt.Fprintf(sum, "%s|%s|%s|", event.EventID, event.EventType, string(event.Payload))
	}
	return hex.EncodeToString(sum.Sum(nil))
}
