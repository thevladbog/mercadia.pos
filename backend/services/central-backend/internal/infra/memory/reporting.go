package memory

import (
	"context"
	"sort"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/domain"
)

func (s *Store) StoreReportingSummary(ctx context.Context, storeID string, window app.ReportingWindow) (app.StoreReportingSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return computeStoreReportingSummary(s, storeID, window), nil
}

func (s *Store) CentralReportingSummary(ctx context.Context, storeIDs []string, window app.ReportingWindow) (app.CentralReportingSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	central := app.CentralReportingSummary{}
	for _, storeID := range storeIDs {
		addStoreSummariesToCentral(&central, computeStoreReportingSummary(s, storeID, window))
	}
	return central, nil
}

func (s *Store) ListStoreReportingSummaries(ctx context.Context, storeIDs []string, window app.ReportingWindow, limit, offset int) ([]app.StoreReportingSummary, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sorted := append([]string(nil), storeIDs...)
	sort.Strings(sorted)
	total := len(sorted)
	if offset >= total {
		return []app.StoreReportingSummary{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	summaries := make([]app.StoreReportingSummary, 0, end-offset)
	for _, storeID := range sorted[offset:end] {
		summaries = append(summaries, computeStoreReportingSummary(s, storeID, window))
	}
	return summaries, total, nil
}

func computeStoreReportingSummary(s *Store, storeID string, window app.ReportingWindow) app.StoreReportingSummary {
	summary := app.StoreReportingSummary{StoreID: storeID}
	for _, document := range s.fiscalDocuments {
		if document.StoreID != storeID || !inReportingWindow(document.FiscalizedAt, window) {
			continue
		}
		switch document.Kind {
		case "receipt":
			summary.FiscalReceiptCount++
			summary.FiscalReceiptAmountMinor += document.AmountMinor
		case "return":
			summary.FiscalReturnCount++
			summary.FiscalReturnAmountMinor += document.AmountMinor
		}
	}
	for _, payment := range s.payments {
		if payment.StoreID != storeID {
			continue
		}
		if payment.Status == domain.SyncedPaymentStatusCaptured && inReportingWindow(payment.CapturedAt, window) {
			summary.PaymentsCapturedAmountMinor += payment.AmountMinor
		}
		if payment.Status == domain.SyncedPaymentStatusCancelled {
			at := payment.UpdatedAt
			if payment.CancelledAt != nil {
				at = *payment.CancelledAt
			}
			if inReportingWindow(at, window) {
				summary.PaymentsCancelledCount++
			}
		}
		if payment.Status == domain.SyncedPaymentStatusRefunded || payment.Status == domain.SyncedPaymentStatusPartiallyRefunded {
			if inReportingWindow(payment.UpdatedAt, window) {
				summary.PaymentsRefundedAmountMinor += payment.RefundedAmountMinor
			}
		}
	}
	for _, ret := range s.returns {
		if ret.StoreID != storeID || !inReportingWindow(ret.SettledAt, window) {
			continue
		}
		summary.ReturnsSettledCount++
		summary.ReturnsSettledAmountMinor += ret.TotalMinor
	}
	for _, movement := range s.cashMovements {
		if movement.StoreID != storeID || !inReportingWindow(movement.PostedAt, window) {
			continue
		}
		summary.CashMovementsPostedCount++
	}
	for _, day := range s.operationalDays {
		if day.StoreID != storeID || !inReportingWindow(day.ClosedAt, window) {
			continue
		}
		summary.OperationalDaysClosedCount++
	}
	return summary
}

func inReportingWindow(t time.Time, window app.ReportingWindow) bool {
	return !t.Before(window.Since) && !t.After(window.Until)
}

func addStoreSummariesToCentral(target *app.CentralReportingSummary, item app.StoreReportingSummary) {
	target.FiscalReceiptCount += item.FiscalReceiptCount
	target.FiscalReceiptAmountMinor += item.FiscalReceiptAmountMinor
	target.FiscalReturnCount += item.FiscalReturnCount
	target.FiscalReturnAmountMinor += item.FiscalReturnAmountMinor
	target.PaymentsCapturedAmountMinor += item.PaymentsCapturedAmountMinor
	target.PaymentsCancelledCount += item.PaymentsCancelledCount
	target.PaymentsRefundedAmountMinor += item.PaymentsRefundedAmountMinor
	target.ReturnsSettledCount += item.ReturnsSettledCount
	target.ReturnsSettledAmountMinor += item.ReturnsSettledAmountMinor
	target.CashMovementsPostedCount += item.CashMovementsPostedCount
	target.OperationalDaysClosedCount += item.OperationalDaysClosedCount
}
