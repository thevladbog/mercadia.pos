package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
)

type Store struct {
	pool *pgxpool.Pool
}

type StoreOption func(*Store) error

func NewStore(ctx context.Context, databaseURL string, options ...StoreOption) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	store := &Store{pool: pool}
	for _, option := range options {
		if err := option(store); err != nil {
			pool.Close()
			return nil, err
		}
	}
	return store, nil
}

func WithProducts(products ...domain.Product) StoreOption {
	return func(store *Store) error {
		ctx := context.Background()
		for _, product := range products {
			if err := store.SaveProduct(ctx, product); err != nil {
				return err
			}
		}
		return nil
	}
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Store) SaveReceipt(ctx context.Context, receipt domain.Receipt) error {
	lines, err := json.Marshal(receipt.Lines)
	if err != nil {
		return fmt.Errorf("marshal receipt lines: %w", err)
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO receipts (
			id, store_id, operational_day_id, business_date, shift_id, terminal_id, cashier_id,
			drawer_id, channel, status, lines, cancel_reason, cancelled_by_id, cancel_approved_by_id,
			cancelled_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13, $14,
			$15, $16, $17
		)
		ON CONFLICT (id) DO UPDATE SET
			store_id = EXCLUDED.store_id,
			operational_day_id = EXCLUDED.operational_day_id,
			business_date = EXCLUDED.business_date,
			shift_id = EXCLUDED.shift_id,
			terminal_id = EXCLUDED.terminal_id,
			cashier_id = EXCLUDED.cashier_id,
			drawer_id = EXCLUDED.drawer_id,
			channel = EXCLUDED.channel,
			status = EXCLUDED.status,
			lines = EXCLUDED.lines,
			cancel_reason = EXCLUDED.cancel_reason,
			cancelled_by_id = EXCLUDED.cancelled_by_id,
			cancel_approved_by_id = EXCLUDED.cancel_approved_by_id,
			cancelled_at = EXCLUDED.cancelled_at,
			updated_at = EXCLUDED.updated_at
	`,
		receipt.ID,
		receipt.StoreID,
		receipt.OperationalDayID,
		receipt.BusinessDate,
		receipt.ShiftID,
		receipt.TerminalID,
		receipt.CashierID,
		receipt.DrawerID,
		receipt.Channel,
		string(receipt.Status),
		lines,
		receipt.CancelReason,
		receipt.CancelledByID,
		receipt.CancelApprovedByID,
		nullTime(receipt.CancelledAt),
		receipt.CreatedAt,
		receipt.UpdatedAt,
	)
	return err
}

func (s *Store) FindReceipt(ctx context.Context, receiptID string) (domain.Receipt, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, store_id, operational_day_id, business_date, shift_id, terminal_id, cashier_id,
			drawer_id, channel, status, lines, cancel_reason, cancelled_by_id, cancel_approved_by_id,
			cancelled_at, created_at, updated_at
		FROM receipts
		WHERE id = $1
	`, receiptID)

	receipt, err := scanReceipt(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Receipt{}, app.ErrReceiptNotFound
	}
	return receipt, err
}

func (s *Store) ListReceiptsByShift(ctx context.Context, shiftID string) ([]domain.Receipt, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, store_id, operational_day_id, business_date, shift_id, terminal_id, cashier_id,
			drawer_id, channel, status, lines, cancel_reason, cancelled_by_id, cancel_approved_by_id,
			cancelled_at, created_at, updated_at
		FROM receipts
		WHERE shift_id = $1
		ORDER BY created_at
	`, shiftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanReceipts(rows)
}

func (s *Store) ListReceiptsByOperationalDay(ctx context.Context, operationalDayID string) ([]domain.Receipt, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, store_id, operational_day_id, business_date, shift_id, terminal_id, cashier_id,
			drawer_id, channel, status, lines, cancel_reason, cancelled_by_id, cancel_approved_by_id,
			cancelled_at, created_at, updated_at
		FROM receipts
		WHERE operational_day_id = $1
		ORDER BY created_at
	`, operationalDayID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanReceipts(rows)
}

func (s *Store) SaveTerminal(ctx context.Context, terminal domain.Terminal) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO terminals (
			id, store_id, kind, status, software_version, last_seen_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			store_id = EXCLUDED.store_id,
			kind = EXCLUDED.kind,
			status = EXCLUDED.status,
			software_version = EXCLUDED.software_version,
			last_seen_at = EXCLUDED.last_seen_at,
			updated_at = EXCLUDED.updated_at
	`,
		terminal.ID,
		terminal.StoreID,
		string(terminal.Kind),
		string(terminal.Status),
		terminal.SoftwareVersion,
		terminal.LastSeenAt,
		terminal.UpdatedAt,
	)
	return err
}

func (s *Store) FindTerminal(ctx context.Context, terminalID string) (domain.Terminal, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, store_id, kind, status, software_version, last_seen_at, updated_at
		FROM terminals
		WHERE id = $1
	`, terminalID)

	terminal, err := scanTerminal(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Terminal{}, app.ErrTerminalNotFound
	}
	return terminal, err
}

func (s *Store) ListTerminalsByStore(ctx context.Context, storeID string) ([]domain.Terminal, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, store_id, kind, status, software_version, last_seen_at, updated_at
		FROM terminals
		WHERE store_id = $1
		ORDER BY id
	`, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	terminals := make([]domain.Terminal, 0)
	for rows.Next() {
		terminal, err := scanTerminal(rows)
		if err != nil {
			return nil, err
		}
		terminals = append(terminals, terminal)
	}
	return terminals, rows.Err()
}

func (s *Store) FindProductByBarcode(ctx context.Context, barcode string) (domain.Product, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, barcodes, unit_price_minor, tax_category_id, active
		FROM products
		WHERE active = TRUE
			AND EXISTS (
				SELECT 1
				FROM jsonb_array_elements_text(barcodes) AS barcode(value)
				WHERE barcode.value = $1
			)
		LIMIT 1
	`, barcode)

	product, err := scanProduct(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Product{}, app.ErrProductNotFound
	}
	return product, err
}

func (s *Store) SaveProduct(ctx context.Context, product domain.Product) error {
	return s.saveProduct(ctx, product)
}

func (s *Store) GetLastSyncedAt(ctx context.Context, storeID string) (time.Time, error) {
	var syncedAt time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT last_synced_at
		FROM catalog_sync_state
		WHERE store_id = $1
	`, storeID).Scan(&syncedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	return syncedAt.UTC(), nil
}

func (s *Store) SaveLastSyncedAt(ctx context.Context, storeID string, syncedAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO catalog_sync_state (store_id, last_synced_at)
		VALUES ($1, $2)
		ON CONFLICT (store_id) DO UPDATE SET
			last_synced_at = EXCLUDED.last_synced_at
	`, storeID, syncedAt.UTC())
	return err
}

func (s *Store) SavePayment(ctx context.Context, payment domain.Payment) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO payments (
			id, receipt_id, method, status, amount_minor, refunded_amount_minor, provider_reference,
			created_at, updated_at, captured_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			receipt_id = EXCLUDED.receipt_id,
			method = EXCLUDED.method,
			status = EXCLUDED.status,
			amount_minor = EXCLUDED.amount_minor,
			refunded_amount_minor = EXCLUDED.refunded_amount_minor,
			provider_reference = EXCLUDED.provider_reference,
			updated_at = EXCLUDED.updated_at,
			captured_at = EXCLUDED.captured_at
	`,
		payment.ID,
		payment.ReceiptID,
		string(payment.Method),
		string(payment.Status),
		payment.AmountMinor,
		payment.RefundedAmountMinor,
		payment.ProviderReference,
		payment.CreatedAt,
		payment.UpdatedAt,
		payment.CapturedAt,
	)
	return err
}

func (s *Store) FindPayment(ctx context.Context, paymentID string) (domain.Payment, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, receipt_id, method, status, amount_minor, refunded_amount_minor, provider_reference,
			created_at, updated_at, captured_at
		FROM payments
		WHERE id = $1
	`, paymentID)
	payment, err := scanPayment(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Payment{}, app.ErrPaymentNotFound
	}
	return payment, err
}

func (s *Store) FindPaymentsByReceipt(ctx context.Context, receiptID string) ([]domain.Payment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, receipt_id, method, status, amount_minor, refunded_amount_minor, provider_reference,
			created_at, updated_at, captured_at
		FROM payments
		WHERE receipt_id = $1
		ORDER BY created_at
	`, receiptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPayments(rows)
}

func (s *Store) CountFiscalizedReceiptsByStoreAndBusinessDate(ctx context.Context, storeID string, businessDate string) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM receipts
		WHERE store_id = $1
			AND status = $2
			AND COALESCE(NULLIF(business_date, ''), to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD')) = $3
	`, storeID, string(domain.ReceiptStatusFiscalized), businessDate).Scan(&count)
	return count, err
}

func (s *Store) ListUnresolvedReceiptsByStoreAndBusinessDate(ctx context.Context, storeID string, businessDate string) ([]domain.Receipt, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, store_id, operational_day_id, business_date, shift_id, terminal_id, cashier_id,
			drawer_id, channel, status, lines, cancel_reason, cancelled_by_id, cancel_approved_by_id,
			cancelled_at, created_at, updated_at
		FROM receipts
		WHERE store_id = $1
			AND COALESCE(NULLIF(business_date, ''), to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD')) = $2
			AND status = ANY($3)
		ORDER BY created_at
	`, storeID, businessDate, []string{
		string(domain.ReceiptStatusDraft),
		string(domain.ReceiptStatusPaymentStarted),
		string(domain.ReceiptStatusPaid),
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanReceipts(rows)
}

func (s *Store) ListUnresolvedReceiptsByShift(ctx context.Context, shiftID string) ([]domain.Receipt, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, store_id, operational_day_id, business_date, shift_id, terminal_id, cashier_id,
			drawer_id, channel, status, lines, cancel_reason, cancelled_by_id, cancel_approved_by_id,
			cancelled_at, created_at, updated_at
		FROM receipts
		WHERE shift_id = $1
			AND status = ANY($2)
		ORDER BY created_at
	`, shiftID, []string{
		string(domain.ReceiptStatusDraft),
		string(domain.ReceiptStatusPaymentStarted),
		string(domain.ReceiptStatusPaid),
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanReceipts(rows)
}

func (s *Store) SaveFiscalDocument(ctx context.Context, document domain.FiscalDocument) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO fiscal_documents (
			id, receipt_id, kind, status, amount_minor, device_id, fiscal_sign, fiscalized_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			receipt_id = EXCLUDED.receipt_id,
			kind = EXCLUDED.kind,
			status = EXCLUDED.status,
			amount_minor = EXCLUDED.amount_minor,
			device_id = EXCLUDED.device_id,
			fiscal_sign = EXCLUDED.fiscal_sign,
			fiscalized_at = EXCLUDED.fiscalized_at
	`,
		document.ID,
		document.ReceiptID,
		string(document.Kind),
		string(document.Status),
		document.AmountMinor,
		document.DeviceID,
		document.FiscalSign,
		document.FiscalizedAt,
		document.CreatedAt,
	)
	return err
}

func (s *Store) FindFiscalDocumentsByReceipt(ctx context.Context, receiptID string) ([]domain.FiscalDocument, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, receipt_id, kind, status, amount_minor, device_id, fiscal_sign, fiscalized_at, created_at
		FROM fiscal_documents
		WHERE receipt_id = $1
		ORDER BY created_at
	`, receiptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFiscalDocuments(rows)
}

func (s *Store) SaveCashMovement(ctx context.Context, movement domain.CashMovement) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO cash_movements (
			id, store_id, type, from_container_id, from_container_type, to_container_id, to_container_type,
			amount_minor, currency, reason, actor_id, approved_by_id, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO UPDATE SET
			store_id = EXCLUDED.store_id,
			type = EXCLUDED.type,
			from_container_id = EXCLUDED.from_container_id,
			from_container_type = EXCLUDED.from_container_type,
			to_container_id = EXCLUDED.to_container_id,
			to_container_type = EXCLUDED.to_container_type,
			amount_minor = EXCLUDED.amount_minor,
			currency = EXCLUDED.currency,
			reason = EXCLUDED.reason,
			actor_id = EXCLUDED.actor_id,
			approved_by_id = EXCLUDED.approved_by_id,
			status = EXCLUDED.status
	`,
		movement.ID,
		movement.StoreID,
		string(movement.Type),
		movement.FromContainerID,
		string(movement.FromContainerType),
		movement.ToContainerID,
		string(movement.ToContainerType),
		movement.AmountMinor,
		movement.Currency,
		movement.Reason,
		movement.ActorID,
		movement.ApprovedByID,
		string(movement.Status),
		movement.CreatedAt,
	)
	return err
}

func (s *Store) ListCashMovements(ctx context.Context, storeID string) ([]domain.CashMovement, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, store_id, type, from_container_id, from_container_type, to_container_id, to_container_type,
			amount_minor, currency, reason, actor_id, approved_by_id, status, created_at
		FROM cash_movements
		WHERE store_id = $1
		ORDER BY created_at
	`, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCashMovements(rows)
}

func (s *Store) SaveCashRecount(ctx context.Context, recount domain.CashRecount) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO cash_recounts (
			id, store_id, business_date, container_id, container_type, currency, expected_minor,
			counted_minor, discrepancy_minor, reason, actor_id, approved_by_id, status, resolution_status,
			resolution_note, resolved_by_id, resolved_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		ON CONFLICT (id) DO UPDATE SET
			store_id = EXCLUDED.store_id,
			business_date = EXCLUDED.business_date,
			container_id = EXCLUDED.container_id,
			container_type = EXCLUDED.container_type,
			currency = EXCLUDED.currency,
			expected_minor = EXCLUDED.expected_minor,
			counted_minor = EXCLUDED.counted_minor,
			discrepancy_minor = EXCLUDED.discrepancy_minor,
			reason = EXCLUDED.reason,
			actor_id = EXCLUDED.actor_id,
			approved_by_id = EXCLUDED.approved_by_id,
			status = EXCLUDED.status,
			resolution_status = EXCLUDED.resolution_status,
			resolution_note = EXCLUDED.resolution_note,
			resolved_by_id = EXCLUDED.resolved_by_id,
			resolved_at = EXCLUDED.resolved_at
	`,
		recount.ID,
		recount.StoreID,
		recount.BusinessDate,
		recount.ContainerID,
		string(recount.ContainerType),
		recount.Currency,
		recount.ExpectedMinor,
		recount.CountedMinor,
		recount.DiscrepancyMinor,
		recount.Reason,
		recount.ActorID,
		recount.ApprovedByID,
		string(recount.Status),
		string(recount.ResolutionStatus),
		recount.ResolutionNote,
		recount.ResolvedByID,
		nullTime(recount.ResolvedAt),
		recount.CreatedAt,
	)
	return err
}

func (s *Store) FindCashRecount(ctx context.Context, recountID string) (domain.CashRecount, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, store_id, business_date, container_id, container_type, currency, expected_minor,
			counted_minor, discrepancy_minor, reason, actor_id, approved_by_id, status, resolution_status,
			resolution_note, resolved_by_id, resolved_at, created_at
		FROM cash_recounts
		WHERE id = $1
	`, recountID)

	recount, err := scanCashRecount(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.CashRecount{}, app.ErrCashRecountNotFound
	}
	return recount, err
}

func (s *Store) ListCashRecounts(ctx context.Context, storeID string) ([]domain.CashRecount, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, store_id, business_date, container_id, container_type, currency, expected_minor,
			counted_minor, discrepancy_minor, reason, actor_id, approved_by_id, status, resolution_status,
			resolution_note, resolved_by_id, resolved_at, created_at
		FROM cash_recounts
		WHERE store_id = $1
		ORDER BY created_at
	`, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCashRecounts(rows)
}

func (s *Store) ListCashRecountsByStoreAndBusinessDate(ctx context.Context, storeID string, businessDate string) ([]domain.CashRecount, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, store_id, business_date, container_id, container_type, currency, expected_minor,
			counted_minor, discrepancy_minor, reason, actor_id, approved_by_id, status, resolution_status,
			resolution_note, resolved_by_id, resolved_at, created_at
		FROM cash_recounts
		WHERE store_id = $1
			AND COALESCE(NULLIF(business_date, ''), to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD')) = $2
		ORDER BY created_at
	`, storeID, businessDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCashRecounts(rows)
}

func (s *Store) ListUnresolvedCashRecountDiscrepanciesByStoreAndBusinessDate(ctx context.Context, storeID string, businessDate string) ([]domain.CashRecount, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, store_id, business_date, container_id, container_type, currency, expected_minor,
			counted_minor, discrepancy_minor, reason, actor_id, approved_by_id, status, resolution_status,
			resolution_note, resolved_by_id, resolved_at, created_at
		FROM cash_recounts
		WHERE store_id = $1
			AND status = $2
			AND resolution_status = $3
			AND COALESCE(NULLIF(business_date, ''), to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD')) = $4
		ORDER BY created_at
	`, storeID, string(domain.CashRecountStatusDiscrepancy), string(domain.CashRecountResolutionStatusOpen), businessDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCashRecounts(rows)
}

func (s *Store) SaveShift(ctx context.Context, shift domain.Shift) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO shifts (
			id, store_id, operational_day_id, business_date, terminal_id, cashier_id, drawer_id,
			status, opening_cash_minor, closing_cash_minor, opened_at, closed_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			store_id = EXCLUDED.store_id,
			operational_day_id = EXCLUDED.operational_day_id,
			business_date = EXCLUDED.business_date,
			terminal_id = EXCLUDED.terminal_id,
			cashier_id = EXCLUDED.cashier_id,
			drawer_id = EXCLUDED.drawer_id,
			status = EXCLUDED.status,
			opening_cash_minor = EXCLUDED.opening_cash_minor,
			closing_cash_minor = EXCLUDED.closing_cash_minor,
			closed_at = EXCLUDED.closed_at,
			updated_at = EXCLUDED.updated_at
	`,
		shift.ID,
		shift.StoreID,
		shift.OperationalDayID,
		shift.BusinessDate,
		shift.TerminalID,
		shift.CashierID,
		shift.DrawerID,
		string(shift.Status),
		shift.OpeningCashMinor,
		shift.ClosingCashMinor,
		shift.OpenedAt,
		nullTime(shift.ClosedAt),
		shift.UpdatedAt,
	)
	return err
}

func (s *Store) FindShift(ctx context.Context, shiftID string) (domain.Shift, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, store_id, operational_day_id, business_date, terminal_id, cashier_id, drawer_id,
			status, opening_cash_minor, closing_cash_minor, opened_at, closed_at, updated_at
		FROM shifts
		WHERE id = $1
	`, shiftID)

	shift, err := scanShift(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Shift{}, app.ErrShiftNotFound
	}
	return shift, err
}

func (s *Store) FindOpenShiftByTerminal(ctx context.Context, terminalID string) (domain.Shift, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, store_id, operational_day_id, business_date, terminal_id, cashier_id, drawer_id,
			status, opening_cash_minor, closing_cash_minor, opened_at, closed_at, updated_at
		FROM shifts
		WHERE terminal_id = $1 AND status = $2
		LIMIT 1
	`, terminalID, string(domain.ShiftStatusOpen))

	shift, err := scanShift(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Shift{}, app.ErrShiftNotFound
	}
	return shift, err
}

func (s *Store) FindOpenShiftByCashier(ctx context.Context, cashierID string) (domain.Shift, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, store_id, operational_day_id, business_date, terminal_id, cashier_id, drawer_id,
			status, opening_cash_minor, closing_cash_minor, opened_at, closed_at, updated_at
		FROM shifts
		WHERE cashier_id = $1 AND status = $2
		LIMIT 1
	`, cashierID, string(domain.ShiftStatusOpen))

	shift, err := scanShift(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Shift{}, app.ErrShiftNotFound
	}
	return shift, err
}

func (s *Store) ListOpenShiftsByStore(ctx context.Context, storeID string) ([]domain.Shift, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, store_id, operational_day_id, business_date, terminal_id, cashier_id, drawer_id,
			status, opening_cash_minor, closing_cash_minor, opened_at, closed_at, updated_at
		FROM shifts
		WHERE store_id = $1 AND status = $2
		ORDER BY opened_at
	`, storeID, string(domain.ShiftStatusOpen))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanShifts(rows)
}

func (s *Store) ListShiftsByOperationalDay(ctx context.Context, operationalDayID string) ([]domain.Shift, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, store_id, operational_day_id, business_date, terminal_id, cashier_id, drawer_id,
			status, opening_cash_minor, closing_cash_minor, opened_at, closed_at, updated_at
		FROM shifts
		WHERE operational_day_id = $1
		ORDER BY opened_at
	`, operationalDayID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanShifts(rows)
}

func (s *Store) SaveOperationalDay(ctx context.Context, day domain.OperationalDay) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO operational_days (
			id, store_id, business_date, status, opened_by_id, closed_by_id, opened_at, closed_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			store_id = EXCLUDED.store_id,
			business_date = EXCLUDED.business_date,
			status = EXCLUDED.status,
			opened_by_id = EXCLUDED.opened_by_id,
			closed_by_id = EXCLUDED.closed_by_id,
			closed_at = EXCLUDED.closed_at,
			updated_at = EXCLUDED.updated_at
	`,
		day.ID,
		day.StoreID,
		day.BusinessDate,
		string(day.Status),
		day.OpenedByID,
		day.ClosedByID,
		day.OpenedAt,
		nullTime(day.ClosedAt),
		day.UpdatedAt,
	)
	return err
}

func (s *Store) FindOperationalDay(ctx context.Context, dayID string) (domain.OperationalDay, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, store_id, business_date, status, opened_by_id, closed_by_id, opened_at, closed_at, updated_at
		FROM operational_days
		WHERE id = $1
	`, dayID)

	day, err := scanOperationalDay(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.OperationalDay{}, app.ErrOperationalDayNotFound
	}
	return day, err
}

func (s *Store) FindOpenOperationalDayByStore(ctx context.Context, storeID string) (domain.OperationalDay, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, store_id, business_date, status, opened_by_id, closed_by_id, opened_at, closed_at, updated_at
		FROM operational_days
		WHERE store_id = $1 AND status = $2
		ORDER BY opened_at
		LIMIT 1
	`, storeID, string(domain.OperationalDayStatusOpen))

	day, err := scanOperationalDay(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.OperationalDay{}, app.ErrOperationalDayNotFound
	}
	return day, err
}

func (s *Store) Find(ctx context.Context, operation string, key string) (app.IdempotencyRecord, bool, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT operation, key, target_id, fingerprint, result, created_at
		FROM idempotency_records
		WHERE operation = $1 AND key = $2
	`, operation, key)

	var record app.IdempotencyRecord
	var resultJSON []byte
	err := row.Scan(
		&record.Operation,
		&record.Key,
		&record.TargetID,
		&record.Fingerprint,
		&resultJSON,
		&record.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return app.IdempotencyRecord{}, false, nil
	}
	if err != nil {
		return app.IdempotencyRecord{}, false, err
	}

	result, err := decodeIdempotencyResult(operation, resultJSON)
	if err != nil {
		return app.IdempotencyRecord{}, false, err
	}
	record.Result = result
	return record, true, nil
}

func (s *Store) Save(ctx context.Context, record app.IdempotencyRecord) error {
	resultJSON, err := json.Marshal(record.Result)
	if err != nil {
		return fmt.Errorf("marshal idempotency result: %w", err)
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO idempotency_records (
			operation, key, target_id, fingerprint, result, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (operation, key) DO UPDATE SET
			target_id = EXCLUDED.target_id,
			fingerprint = EXCLUDED.fingerprint,
			result = EXCLUDED.result,
			created_at = EXCLUDED.created_at
	`,
		record.Operation,
		record.Key,
		record.TargetID,
		record.Fingerprint,
		resultJSON,
		record.CreatedAt,
	)
	return err
}

func (s *Store) saveProduct(ctx context.Context, product domain.Product) error {
	product = cloneProduct(product)
	barcodes, err := json.Marshal(product.Barcodes)
	if err != nil {
		return fmt.Errorf("marshal product barcodes: %w", err)
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO products (
			id, name, barcodes, unit_price_minor, tax_category_id, active
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			barcodes = EXCLUDED.barcodes,
			unit_price_minor = EXCLUDED.unit_price_minor,
			tax_category_id = EXCLUDED.tax_category_id,
			active = EXCLUDED.active
	`,
		product.ID,
		product.Name,
		barcodes,
		product.UnitPriceMinor,
		product.TaxCategoryID,
		product.Active,
	)
	return err
}

func cloneProduct(product domain.Product) domain.Product {
	product.Barcodes = append([]string(nil), product.Barcodes...)
	return product
}

func decodeIdempotencyResult(operation string, data []byte) (any, error) {
	switch {
	case strings.HasPrefix(operation, "checkout."):
		var result app.ReceiptResult
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return result, nil
	case strings.HasPrefix(operation, "payments."):
		var result app.PaymentResult
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return result, nil
	case strings.HasPrefix(operation, "fiscalization."):
		var result app.FiscalDocumentResult
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return result, nil
	case strings.HasPrefix(operation, "shifts."):
		var result app.ShiftResult
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return result, nil
	case strings.HasPrefix(operation, "cash."):
		switch operation {
		case "cash.create_cash_movement":
			var result app.CashMovementResult
			if err := json.Unmarshal(data, &result); err != nil {
				return nil, err
			}
			return result, nil
		default:
			var result app.CashRecountResult
			if err := json.Unmarshal(data, &result); err != nil {
				return nil, err
			}
			return result, nil
		}
	case strings.HasPrefix(operation, "operational_days."):
		var result app.OperationalDayResult
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return result, nil
	case strings.HasPrefix(operation, "terminals."):
		var result app.TerminalResult
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unknown idempotency operation %q", operation)
	}
}

func nullTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanReceipt(row rowScanner) (domain.Receipt, error) {
	var receipt domain.Receipt
	var status string
	var linesJSON []byte
	var cancelledAt *time.Time

	err := row.Scan(
		&receipt.ID,
		&receipt.StoreID,
		&receipt.OperationalDayID,
		&receipt.BusinessDate,
		&receipt.ShiftID,
		&receipt.TerminalID,
		&receipt.CashierID,
		&receipt.DrawerID,
		&receipt.Channel,
		&status,
		&linesJSON,
		&receipt.CancelReason,
		&receipt.CancelledByID,
		&receipt.CancelApprovedByID,
		&cancelledAt,
		&receipt.CreatedAt,
		&receipt.UpdatedAt,
	)
	if err != nil {
		return domain.Receipt{}, err
	}

	receipt.Status = domain.ReceiptStatus(status)
	if len(linesJSON) > 0 {
		if err := json.Unmarshal(linesJSON, &receipt.Lines); err != nil {
			return domain.Receipt{}, err
		}
	}
	if cancelledAt != nil {
		receipt.CancelledAt = *cancelledAt
	}
	return receipt, nil
}

func scanReceipts(rows pgx.Rows) ([]domain.Receipt, error) {
	receipts := []domain.Receipt{}
	for rows.Next() {
		receipt, err := scanReceipt(rows)
		if err != nil {
			return nil, err
		}
		receipts = append(receipts, receipt)
	}
	return receipts, rows.Err()
}

func scanTerminal(row rowScanner) (domain.Terminal, error) {
	var terminal domain.Terminal
	var kind string
	var status string

	err := row.Scan(
		&terminal.ID,
		&terminal.StoreID,
		&kind,
		&status,
		&terminal.SoftwareVersion,
		&terminal.LastSeenAt,
		&terminal.UpdatedAt,
	)
	if err != nil {
		return domain.Terminal{}, err
	}

	terminal.Kind = domain.TerminalKind(kind)
	terminal.Status = domain.TerminalStatus(status)
	return terminal, nil
}

func scanProduct(row rowScanner) (domain.Product, error) {
	var product domain.Product
	var barcodesJSON []byte

	err := row.Scan(
		&product.ID,
		&product.Name,
		&barcodesJSON,
		&product.UnitPriceMinor,
		&product.TaxCategoryID,
		&product.Active,
	)
	if err != nil {
		return domain.Product{}, err
	}

	if len(barcodesJSON) > 0 {
		if err := json.Unmarshal(barcodesJSON, &product.Barcodes); err != nil {
			return domain.Product{}, err
		}
	}
	return product, nil
}

func scanPayment(row rowScanner) (domain.Payment, error) {
	var payment domain.Payment
	var method string
	var status string

	err := row.Scan(
		&payment.ID,
		&payment.ReceiptID,
		&method,
		&status,
		&payment.AmountMinor,
		&payment.RefundedAmountMinor,
		&payment.ProviderReference,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.CapturedAt,
	)
	if err != nil {
		return domain.Payment{}, err
	}

	payment.Method = domain.PaymentMethod(method)
	payment.Status = domain.PaymentStatus(status)
	return payment, nil
}

func scanPayments(rows pgx.Rows) ([]domain.Payment, error) {
	payments := []domain.Payment{}
	for rows.Next() {
		payment, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}
	return payments, rows.Err()
}

func scanFiscalDocument(row rowScanner) (domain.FiscalDocument, error) {
	var document domain.FiscalDocument
	var kind string
	var status string

	err := row.Scan(
		&document.ID,
		&document.ReceiptID,
		&kind,
		&status,
		&document.AmountMinor,
		&document.DeviceID,
		&document.FiscalSign,
		&document.FiscalizedAt,
		&document.CreatedAt,
	)
	if err != nil {
		return domain.FiscalDocument{}, err
	}

	document.Kind = domain.FiscalDocumentKind(kind)
	document.Status = domain.FiscalDocumentStatus(status)
	return document, nil
}

func scanFiscalDocuments(rows pgx.Rows) ([]domain.FiscalDocument, error) {
	documents := []domain.FiscalDocument{}
	for rows.Next() {
		document, err := scanFiscalDocument(rows)
		if err != nil {
			return nil, err
		}
		documents = append(documents, document)
	}
	return documents, rows.Err()
}

func scanCashMovement(row rowScanner) (domain.CashMovement, error) {
	var movement domain.CashMovement
	var movementType string
	var fromContainerType string
	var toContainerType string
	var status string

	err := row.Scan(
		&movement.ID,
		&movement.StoreID,
		&movementType,
		&movement.FromContainerID,
		&fromContainerType,
		&movement.ToContainerID,
		&toContainerType,
		&movement.AmountMinor,
		&movement.Currency,
		&movement.Reason,
		&movement.ActorID,
		&movement.ApprovedByID,
		&status,
		&movement.CreatedAt,
	)
	if err != nil {
		return domain.CashMovement{}, err
	}

	movement.Type = domain.CashMovementType(movementType)
	movement.FromContainerType = domain.CashContainerType(fromContainerType)
	movement.ToContainerType = domain.CashContainerType(toContainerType)
	movement.Status = domain.CashMovementStatus(status)
	return movement, nil
}

func scanCashMovements(rows pgx.Rows) ([]domain.CashMovement, error) {
	movements := []domain.CashMovement{}
	for rows.Next() {
		movement, err := scanCashMovement(rows)
		if err != nil {
			return nil, err
		}
		movements = append(movements, movement)
	}
	return movements, rows.Err()
}

func scanCashRecount(row rowScanner) (domain.CashRecount, error) {
	var recount domain.CashRecount
	var containerType string
	var status string
	var resolutionStatus string
	var resolvedAt *time.Time

	err := row.Scan(
		&recount.ID,
		&recount.StoreID,
		&recount.BusinessDate,
		&recount.ContainerID,
		&containerType,
		&recount.Currency,
		&recount.ExpectedMinor,
		&recount.CountedMinor,
		&recount.DiscrepancyMinor,
		&recount.Reason,
		&recount.ActorID,
		&recount.ApprovedByID,
		&status,
		&resolutionStatus,
		&recount.ResolutionNote,
		&recount.ResolvedByID,
		&resolvedAt,
		&recount.CreatedAt,
	)
	if err != nil {
		return domain.CashRecount{}, err
	}

	recount.ContainerType = domain.CashContainerType(containerType)
	recount.Status = domain.CashRecountStatus(status)
	recount.ResolutionStatus = domain.CashRecountResolutionStatus(resolutionStatus)
	if resolvedAt != nil {
		recount.ResolvedAt = *resolvedAt
	}
	return recount, nil
}

func scanCashRecounts(rows pgx.Rows) ([]domain.CashRecount, error) {
	recounts := []domain.CashRecount{}
	for rows.Next() {
		recount, err := scanCashRecount(rows)
		if err != nil {
			return nil, err
		}
		recounts = append(recounts, recount)
	}
	return recounts, rows.Err()
}

func scanShift(row rowScanner) (domain.Shift, error) {
	var shift domain.Shift
	var status string
	var closedAt *time.Time

	err := row.Scan(
		&shift.ID,
		&shift.StoreID,
		&shift.OperationalDayID,
		&shift.BusinessDate,
		&shift.TerminalID,
		&shift.CashierID,
		&shift.DrawerID,
		&status,
		&shift.OpeningCashMinor,
		&shift.ClosingCashMinor,
		&shift.OpenedAt,
		&closedAt,
		&shift.UpdatedAt,
	)
	if err != nil {
		return domain.Shift{}, err
	}

	shift.Status = domain.ShiftStatus(status)
	if closedAt != nil {
		shift.ClosedAt = *closedAt
	}
	return shift, nil
}

func scanShifts(rows pgx.Rows) ([]domain.Shift, error) {
	shifts := []domain.Shift{}
	for rows.Next() {
		shift, err := scanShift(rows)
		if err != nil {
			return nil, err
		}
		shifts = append(shifts, shift)
	}
	return shifts, rows.Err()
}

func scanOperationalDay(row rowScanner) (domain.OperationalDay, error) {
	var day domain.OperationalDay
	var status string
	var closedAt *time.Time

	err := row.Scan(
		&day.ID,
		&day.StoreID,
		&day.BusinessDate,
		&status,
		&day.OpenedByID,
		&day.ClosedByID,
		&day.OpenedAt,
		&closedAt,
		&day.UpdatedAt,
	)
	if err != nil {
		return domain.OperationalDay{}, err
	}

	day.Status = domain.OperationalDayStatus(status)
	if closedAt != nil {
		day.ClosedAt = *closedAt
	}
	return day, nil
}

func (s *Store) SaveOutboxEvent(ctx context.Context, event domain.OutboxEvent) error {
	event, err := domain.NewOutboxEvent(event)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO outbox_events (
			id, aggregate_type, aggregate_id, event_type, payload, created_at, published_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		ON CONFLICT (id) DO NOTHING
	`,
		event.ID,
		event.AggregateType,
		event.AggregateID,
		event.EventType,
		event.Payload,
		event.CreatedAt,
		event.PublishedAt,
	)
	if err != nil {
		return fmt.Errorf("save outbox event: %w", err)
	}
	return nil
}

func (s *Store) ListPendingOutboxEvents(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, created_at, published_at
		FROM outbox_events
		WHERE published_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list pending outbox events: %w", err)
	}
	defer rows.Close()

	events := []domain.OutboxEvent{}
	for rows.Next() {
		event, err := scanOutboxEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *Store) MarkOutboxEventPublished(ctx context.Context, eventID string, publishedAt time.Time) (bool, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE outbox_events
		SET published_at = $2
		WHERE id = $1 AND published_at IS NULL
	`, eventID, publishedAt.UTC())
	if err != nil {
		return false, fmt.Errorf("mark outbox event published: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) CountOutboxEvents(ctx context.Context) (pending int64, published int64, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE published_at IS NULL),
			COUNT(*) FILTER (WHERE published_at IS NOT NULL)
		FROM outbox_events
	`).Scan(&pending, &published)
	if err != nil {
		return 0, 0, fmt.Errorf("count outbox events: %w", err)
	}
	return pending, published, nil
}

func scanOutboxEvent(row rowScanner) (domain.OutboxEvent, error) {
	var event domain.OutboxEvent
	var publishedAt *time.Time

	err := row.Scan(
		&event.ID,
		&event.AggregateType,
		&event.AggregateID,
		&event.EventType,
		&event.Payload,
		&event.CreatedAt,
		&publishedAt,
	)
	if err != nil {
		return domain.OutboxEvent{}, err
	}
	event.PublishedAt = publishedAt
	return event, nil
}
