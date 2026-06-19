package app

import (
	"context"
	"errors"

	"mercadia.dev/pos/services/central-backend/internal/domain"
)

var (
	ErrFiscalDocumentNotFound    = errors.New("fiscal document not found")
	ErrInvalidFiscalDocumentQuery = errors.New("invalid fiscal document query")
)

type SyncedFiscalDocumentRepository interface {
	SaveFiscalDocument(ctx context.Context, document domain.SyncedFiscalDocument) error
	FindFiscalDocument(ctx context.Context, storeID string, fiscalDocumentID string) (domain.SyncedFiscalDocument, error)
	ListFiscalDocuments(ctx context.Context, storeID string, limit, offset int) ([]domain.SyncedFiscalDocument, int, error)
}

type FiscalDocumentsService struct {
	stores          StoreRepository
	fiscalDocuments SyncedFiscalDocumentRepository
}

func NewFiscalDocumentsService(stores StoreRepository, fiscalDocuments SyncedFiscalDocumentRepository) *FiscalDocumentsService {
	return &FiscalDocumentsService{
		stores:          stores,
		fiscalDocuments: fiscalDocuments,
	}
}

func (s *FiscalDocumentsService) ListFiscalDocuments(ctx context.Context, storeID string, params PageParams) (PageResult[domain.SyncedFiscalDocument], error) {
	if storeID == "" {
		return PageResult[domain.SyncedFiscalDocument]{}, ErrInvalidFiscalDocumentQuery
	}
	if _, err := s.stores.FindStore(ctx, storeID); err != nil {
		return PageResult[domain.SyncedFiscalDocument]{}, err
	}
	documents, total, err := s.fiscalDocuments.ListFiscalDocuments(ctx, storeID, params.Limit, params.Offset)
	if err != nil {
		return PageResult[domain.SyncedFiscalDocument]{}, err
	}
	return PageResult[domain.SyncedFiscalDocument]{Items: documents, TotalCount: total}, nil
}

func (s *FiscalDocumentsService) GetFiscalDocument(ctx context.Context, storeID string, fiscalDocumentID string) (domain.SyncedFiscalDocument, error) {
	if storeID == "" || fiscalDocumentID == "" {
		return domain.SyncedFiscalDocument{}, ErrInvalidFiscalDocumentQuery
	}
	if _, err := s.stores.FindStore(ctx, storeID); err != nil {
		return domain.SyncedFiscalDocument{}, err
	}
	return s.fiscalDocuments.FindFiscalDocument(ctx, storeID, fiscalDocumentID)
}
