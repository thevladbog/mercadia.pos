package app

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidReportingQuery = errors.New("invalid reporting query")
)

type ReportingWindow struct {
	Since time.Time
	Until time.Time
}

type StoreReportingSummary struct {
	StoreID                     string
	Since                       time.Time
	Until                       time.Time
	FiscalReceiptCount          int
	FiscalReceiptAmountMinor    int64
	FiscalReturnCount           int
	FiscalReturnAmountMinor     int64
	PaymentsCapturedAmountMinor int64
	PaymentsCancelledCount      int
	PaymentsRefundedAmountMinor int64
	ReturnsSettledCount         int
	ReturnsSettledAmountMinor   int64
	CashMovementsPostedCount    int
	OperationalDaysClosedCount  int
}

type CentralReportingSummary struct {
	Region                      string
	Since                       time.Time
	Until                       time.Time
	StoreCount                  int
	FiscalReceiptCount          int
	FiscalReceiptAmountMinor    int64
	FiscalReturnCount           int
	FiscalReturnAmountMinor     int64
	PaymentsCapturedAmountMinor int64
	PaymentsCancelledCount      int
	PaymentsRefundedAmountMinor int64
	ReturnsSettledCount         int
	ReturnsSettledAmountMinor   int64
	CashMovementsPostedCount    int
	OperationalDaysClosedCount  int
}

type ReportingRepository interface {
	StoreReportingSummary(ctx context.Context, storeID string, window ReportingWindow) (StoreReportingSummary, error)
	CentralReportingSummary(ctx context.Context, storeIDs []string, window ReportingWindow) (CentralReportingSummary, error)
	ListStoreReportingSummaries(ctx context.Context, storeIDs []string, window ReportingWindow, limit, offset int) ([]StoreReportingSummary, int, error)
}

type ReportingService struct {
	stores    StoreRepository
	reporting ReportingRepository
}

func NewReportingService(stores StoreRepository, reporting ReportingRepository) *ReportingService {
	return &ReportingService{
		stores:    stores,
		reporting: reporting,
	}
}

func ParseReportingWindow(sinceRaw string, untilRaw string) (ReportingWindow, error) {
	if sinceRaw == "" || untilRaw == "" {
		return ReportingWindow{}, ErrInvalidReportingQuery
	}
	since, err := time.Parse(time.RFC3339, sinceRaw)
	if err != nil {
		return ReportingWindow{}, ErrInvalidReportingQuery
	}
	until, err := time.Parse(time.RFC3339, untilRaw)
	if err != nil {
		return ReportingWindow{}, ErrInvalidReportingQuery
	}
	since = since.UTC()
	until = until.UTC()
	if until.Before(since) {
		return ReportingWindow{}, ErrInvalidReportingQuery
	}
	return ReportingWindow{Since: since, Until: until}, nil
}

func (s *ReportingService) GetStoreSummary(ctx context.Context, storeID string, window ReportingWindow) (StoreReportingSummary, error) {
	if storeID == "" {
		return StoreReportingSummary{}, ErrInvalidReportingQuery
	}
	if _, err := s.stores.FindStore(ctx, storeID); err != nil {
		return StoreReportingSummary{}, err
	}
	summary, err := s.reporting.StoreReportingSummary(ctx, storeID, window)
	if err != nil {
		return StoreReportingSummary{}, err
	}
	summary.Since = window.Since
	summary.Until = window.Until
	return summary, nil
}

func (s *ReportingService) GetCentralSummary(ctx context.Context, window ReportingWindow, region string) (CentralReportingSummary, error) {
	storeIDs, err := s.storeIDsForRegion(ctx, region)
	if err != nil {
		return CentralReportingSummary{}, err
	}
	summary, err := s.reporting.CentralReportingSummary(ctx, storeIDs, window)
	if err != nil {
		return CentralReportingSummary{}, err
	}
	summary.Region = region
	summary.Since = window.Since
	summary.Until = window.Until
	summary.StoreCount = len(storeIDs)
	return summary, nil
}

func (s *ReportingService) ListStoreSummaries(ctx context.Context, window ReportingWindow, region string, params PageParams) (PageResult[StoreReportingSummary], error) {
	storeIDs, err := s.storeIDsForRegion(ctx, region)
	if err != nil {
		return PageResult[StoreReportingSummary]{}, err
	}
	summaries, total, err := s.reporting.ListStoreReportingSummaries(ctx, storeIDs, window, params.Limit, params.Offset)
	if err != nil {
		return PageResult[StoreReportingSummary]{}, err
	}
	for i := range summaries {
		summaries[i].Since = window.Since
		summaries[i].Until = window.Until
	}
	return PageResult[StoreReportingSummary]{Items: summaries, TotalCount: total}, nil
}

func (s *ReportingService) storeIDsForRegion(ctx context.Context, region string) ([]string, error) {
	stores, err := s.stores.ListStores(ctx)
	if err != nil {
		return nil, err
	}
	storeIDs := make([]string, 0, len(stores))
	for _, store := range stores {
		if region != "" && store.Region != region {
			continue
		}
		storeIDs = append(storeIDs, store.ID)
	}
	return storeIDs, nil
}
