package app

import (
	"context"
	"errors"
	"sort"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

type CashBalanceProvider interface {
	ListCashBalances(ctx context.Context, storeID string) ([]domain.CashBalance, error)
}

type TerminalOverview struct {
	Terminal                 domain.Terminal
	ShiftID                  string
	CashierID                string
	DrawerID                 string
	ReceiptCount             int
	RevenueMinor             int64
	DrawerBalanceMinor       int64
	CurrentReceiptID         string
	CurrentReceiptStatus     domain.ReceiptStatus
	CurrentReceiptTotalMinor int64
	AttentionNeeded          bool
}

type StoreMonitoringSummary struct {
	RevenueMinorToday      int64
	DrawerCashMinor        int64
	ActiveTerminalCount    int
	FreeTerminalCount      int
	OfflineTerminalCount   int
	AttentionTerminalCount int
	ReceiptCountToday      int
	AverageReceiptMinor    int64
}

type TerminalMonitoringService struct {
	terminals    TerminalRepository
	shifts       ShiftRepository
	receipts     ReceiptRepository
	days         OperationalDayRepository
	cash         CashBalanceProvider
	now          func() time.Time
	offlineAfter time.Duration
}

type TerminalMonitoringOption func(*TerminalMonitoringService)

func NewTerminalMonitoringService(
	terminals TerminalRepository,
	shifts ShiftRepository,
	receipts ReceiptRepository,
	days OperationalDayRepository,
	cash CashBalanceProvider,
	options ...TerminalMonitoringOption,
) *TerminalMonitoringService {
	service := &TerminalMonitoringService{
		terminals: terminals,
		shifts:    shifts,
		receipts:  receipts,
		days:      days,
		cash:      cash,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithTerminalMonitoringClock(now func() time.Time) TerminalMonitoringOption {
	return func(service *TerminalMonitoringService) {
		service.now = now
	}
}

func WithTerminalMonitoringOfflineAfter(duration time.Duration) TerminalMonitoringOption {
	return func(service *TerminalMonitoringService) {
		service.offlineAfter = duration
	}
}

func (s *TerminalMonitoringService) ListTerminalOverviews(ctx context.Context, storeID string, params PageParams) (PageResult[TerminalOverview], error) {
	if storeID == "" {
		return PageResult[TerminalOverview]{}, ErrInvalidTerminalCommand
	}

	overviews, err := s.buildTerminalOverviews(ctx, storeID)
	if err != nil {
		return PageResult[TerminalOverview]{}, err
	}
	return PaginateSlice(overviews, params), nil
}

func (s *TerminalMonitoringService) GetStoreMonitoringSummary(ctx context.Context, storeID string) (StoreMonitoringSummary, error) {
	if storeID == "" {
		return StoreMonitoringSummary{}, ErrInvalidTerminalCommand
	}

	overviews, err := s.buildTerminalOverviews(ctx, storeID)
	if err != nil {
		return StoreMonitoringSummary{}, err
	}

	summary := StoreMonitoringSummary{}
	for _, overview := range overviews {
		if overview.AttentionNeeded {
			summary.AttentionTerminalCount++
		}
		if overview.ShiftID != "" {
			summary.ActiveTerminalCount++
			continue
		}
		if overview.Terminal.Status == domain.TerminalStatusOffline {
			summary.OfflineTerminalCount++
			continue
		}
		summary.FreeTerminalCount++
	}

	balances, err := s.cash.ListCashBalances(ctx, storeID)
	if err != nil {
		return StoreMonitoringSummary{}, err
	}
	for _, balance := range balances {
		if balance.ContainerType == domain.CashContainerTypeDrawer {
			summary.DrawerCashMinor += balance.BalanceMinor
		}
	}

	day, err := s.days.FindOpenOperationalDayByStore(ctx, storeID)
	if err != nil {
		if errors.Is(err, ErrOperationalDayNotFound) {
			return summary, nil
		}
		return StoreMonitoringSummary{}, err
	}

	dayReceipts, err := s.receipts.ListReceiptsByOperationalDay(ctx, day.ID)
	if err != nil {
		return StoreMonitoringSummary{}, err
	}
	var revenueTotal int64
	for _, receipt := range dayReceipts {
		if receipt.Status != domain.ReceiptStatusFiscalized {
			continue
		}
		summary.ReceiptCountToday++
		revenueTotal += receipt.TotalMinor()
	}
	summary.RevenueMinorToday = revenueTotal
	if summary.ReceiptCountToday > 0 {
		summary.AverageReceiptMinor = revenueTotal / int64(summary.ReceiptCountToday)
	}

	return summary, nil
}

func (s *TerminalMonitoringService) buildTerminalOverviews(ctx context.Context, storeID string) ([]TerminalOverview, error) {
	terminals, err := s.terminals.ListTerminalsByStore(ctx, storeID)
	if err != nil {
		return nil, err
	}
	sort.Slice(terminals, func(i, j int) bool {
		return terminals[i].ID < terminals[j].ID
	})

	openShifts, err := s.shifts.ListOpenShiftsByStore(ctx, storeID)
	if err != nil {
		return nil, err
	}
	shiftsByTerminal := map[string]domain.Shift{}
	for _, shift := range openShifts {
		shiftsByTerminal[shift.TerminalID] = shift
	}

	balances, err := s.cash.ListCashBalances(ctx, storeID)
	if err != nil {
		return nil, err
	}
	drawerBalances := map[string]int64{}
	for _, balance := range balances {
		if balance.ContainerType == domain.CashContainerTypeDrawer {
			drawerBalances[balance.ContainerID] = balance.BalanceMinor
		}
	}

	overviews := make([]TerminalOverview, 0, len(terminals))
	for _, terminal := range terminals {
		terminal.Status = DeriveTerminalListStatus(terminal, s.now(), s.offlineAfter)
		overview := TerminalOverview{Terminal: terminal}

		if shift, ok := shiftsByTerminal[terminal.ID]; ok {
			overview.ShiftID = shift.ID
			overview.CashierID = shift.CashierID
			overview.DrawerID = shift.DrawerID
			overview.DrawerBalanceMinor = drawerBalances[shift.DrawerID]

			receipts, err := s.receipts.ListReceiptsByShift(ctx, shift.ID)
			if err != nil {
				return nil, err
			}
			applyShiftReceiptStats(&overview, receipts)
		}

		overview.AttentionNeeded = overview.Terminal.Status == domain.TerminalStatusOffline ||
			overview.CurrentReceiptID != ""
		overviews = append(overviews, overview)
	}

	return overviews, nil
}

func applyShiftReceiptStats(overview *TerminalOverview, receipts []domain.Receipt) {
	var currentReceipt *domain.Receipt
	for _, receipt := range receipts {
		if receipt.Status == domain.ReceiptStatusFiscalized {
			overview.ReceiptCount++
			overview.RevenueMinor += receipt.TotalMinor()
		}
		if receipt.Status != domain.ReceiptStatusDraft && receipt.Status != domain.ReceiptStatusPaymentStarted {
			continue
		}
		if currentReceipt == nil || receipt.UpdatedAt.After(currentReceipt.UpdatedAt) {
			copyReceipt := receipt
			currentReceipt = &copyReceipt
		}
	}
	if currentReceipt != nil {
		overview.CurrentReceiptID = currentReceipt.ID
		overview.CurrentReceiptStatus = currentReceipt.Status
		overview.CurrentReceiptTotalMinor = currentReceipt.TotalMinor()
	}
}
