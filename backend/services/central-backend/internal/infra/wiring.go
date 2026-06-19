package infra

import (
	"context"
	"fmt"
	"os"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/infra/memory"
	"mercadia.dev/pos/services/central-backend/internal/infra/postgres"
)

type Repository interface {
	app.StoreRepository
	app.SyncEventRepository
	app.CatalogProductRepository
	app.SyncedPaymentRepository
	app.SyncedCashMovementRepository
	app.SyncedFiscalDocumentRepository
	app.IdempotencyStore
}

type Handle struct {
	repo   Repository
	closer func()
}

func (h Handle) Repository() Repository {
	return h.repo
}

func (h Handle) Close() {
	if h.closer != nil {
		h.closer()
	}
}

func NewHandle(ctx context.Context) (Handle, error) {
	databaseURL := os.Getenv("MERCADIA_CENTRAL_BACKEND_DATABASE_URL")
	if databaseURL == "" {
		store := memory.NewStore()
		return Handle{repo: store}, nil
	}

	store, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		return Handle{}, fmt.Errorf("open postgres repository: %w", err)
	}
	return Handle{
		repo: store,
		closer: func() {
			store.Close()
		},
	}, nil
}
