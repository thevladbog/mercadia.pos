package postgres

import (
	"context"
	"sort"

	"mercadia.dev/pos/services/central-backend/internal/app"
)

func (s *Store) StoreReportingSummary(ctx context.Context, storeID string, window app.ReportingWindow) (app.StoreReportingSummary, error) {
	return s.aggregateStoreReporting(ctx, storeID, window)
}

func (s *Store) CentralReportingSummary(ctx context.Context, storeIDs []string, window app.ReportingWindow) (app.CentralReportingSummary, error) {
	if len(storeIDs) == 0 {
		return app.CentralReportingSummary{}, nil
	}
	central := app.CentralReportingSummary{}
	for _, storeID := range storeIDs {
		summary, err := s.aggregateStoreReporting(ctx, storeID, window)
		if err != nil {
			return app.CentralReportingSummary{}, err
		}
		addCentralTotals(&central, summary)
	}
	return central, nil
}

func (s *Store) ListStoreReportingSummaries(ctx context.Context, storeIDs []string, window app.ReportingWindow, limit, offset int) ([]app.StoreReportingSummary, int, error) {
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
		summary, err := s.aggregateStoreReporting(ctx, storeID, window)
		if err != nil {
			return nil, 0, err
		}
		summaries = append(summaries, summary)
	}
	return summaries, total, nil
}

func (s *Store) aggregateStoreReporting(ctx context.Context, storeID string, window app.ReportingWindow) (app.StoreReportingSummary, error) {
	summary := app.StoreReportingSummary{StoreID: storeID}

	row := s.pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(amount_minor), 0)
		FROM synced_fiscal_documents
		WHERE store_id = $1 AND kind = 'receipt'
			AND fiscalized_at >= $2 AND fiscalized_at <= $3
	`, storeID, window.Since, window.Until)
	if err := row.Scan(&summary.FiscalReceiptCount, &summary.FiscalReceiptAmountMinor); err != nil {
		return app.StoreReportingSummary{}, err
	}

	row = s.pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(amount_minor), 0)
		FROM synced_fiscal_documents
		WHERE store_id = $1 AND kind = 'return'
			AND fiscalized_at >= $2 AND fiscalized_at <= $3
	`, storeID, window.Since, window.Until)
	if err := row.Scan(&summary.FiscalReturnCount, &summary.FiscalReturnAmountMinor); err != nil {
		return app.StoreReportingSummary{}, err
	}

	row = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_minor), 0)
		FROM synced_payments
		WHERE store_id = $1 AND status = 'captured'
			AND captured_at >= $2 AND captured_at <= $3
	`, storeID, window.Since, window.Until)
	if err := row.Scan(&summary.PaymentsCapturedAmountMinor); err != nil {
		return app.StoreReportingSummary{}, err
	}

	row = s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM synced_payments
		WHERE store_id = $1 AND status = 'cancelled'
			AND COALESCE(cancelled_at, updated_at) >= $2
			AND COALESCE(cancelled_at, updated_at) <= $3
	`, storeID, window.Since, window.Until)
	if err := row.Scan(&summary.PaymentsCancelledCount); err != nil {
		return app.StoreReportingSummary{}, err
	}

	row = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(refunded_amount_minor), 0)
		FROM synced_payments
		WHERE store_id = $1 AND status IN ('refunded', 'partially_refunded')
			AND updated_at >= $2 AND updated_at <= $3
	`, storeID, window.Since, window.Until)
	if err := row.Scan(&summary.PaymentsRefundedAmountMinor); err != nil {
		return app.StoreReportingSummary{}, err
	}

	row = s.pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(total_minor), 0)
		FROM synced_returns
		WHERE store_id = $1 AND settled_at >= $2 AND settled_at <= $3
	`, storeID, window.Since, window.Until)
	if err := row.Scan(&summary.ReturnsSettledCount, &summary.ReturnsSettledAmountMinor); err != nil {
		return app.StoreReportingSummary{}, err
	}

	row = s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM synced_cash_movements
		WHERE store_id = $1 AND posted_at >= $2 AND posted_at <= $3
	`, storeID, window.Since, window.Until)
	if err := row.Scan(&summary.CashMovementsPostedCount); err != nil {
		return app.StoreReportingSummary{}, err
	}

	row = s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM synced_operational_days
		WHERE store_id = $1 AND closed_at >= $2 AND closed_at <= $3
	`, storeID, window.Since, window.Until)
	if err := row.Scan(&summary.OperationalDaysClosedCount); err != nil {
		return app.StoreReportingSummary{}, err
	}

	return summary, nil
}

func addCentralTotals(target *app.CentralReportingSummary, item app.StoreReportingSummary) {
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
