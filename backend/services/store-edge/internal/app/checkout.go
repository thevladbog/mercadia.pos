package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

var (
	ErrReceiptNotFound            = errors.New("receipt not found")
	ErrIdempotencyKeyReused       = errors.New("idempotency key reused for different command")
	ErrIdempotencyKeyRequired     = errors.New("idempotency key is required")
	ErrInvalidCheckoutCommand     = errors.New("invalid checkout command")
	ErrIdempotencyResultMissing   = errors.New("idempotency result missing")
	ErrOpenShiftRequired          = errors.New("open cashier shift is required")
	ErrOpenOperationalDayRequired = errors.New("open operational day is required")
	ErrReceiptCannotBeCancelled   = errors.New("receipt cannot be cancelled")
)

type ReceiptRepository interface {
	SaveReceipt(ctx context.Context, receipt domain.Receipt) error
	FindReceipt(ctx context.Context, receiptID string) (domain.Receipt, error)
	ListReceiptsByShift(ctx context.Context, shiftID string) ([]domain.Receipt, error)
	ListReceiptsByOperationalDay(ctx context.Context, operationalDayID string) ([]domain.Receipt, error)
}

type IdempotencyStore interface {
	Find(ctx context.Context, operation string, key string) (IdempotencyRecord, bool, error)
	Save(ctx context.Context, record IdempotencyRecord) error
}

type IdempotencyRecord struct {
	Operation   string
	Key         string
	TargetID    string
	Fingerprint string
	Result      any
	CreatedAt   time.Time
}

type CheckoutService struct {
	receipts    ReceiptRepository
	idempotency IdempotencyStore
	products    ProductRepository
	shifts      ShiftRepository
	days        OperationalDayRepository
	now         func() time.Time
	newID       func(prefix string) string
}

type CheckoutOption func(*CheckoutService)

func NewCheckoutService(receipts ReceiptRepository, idempotency IdempotencyStore, options ...CheckoutOption) *CheckoutService {
	service := &CheckoutService{
		receipts:    receipts,
		idempotency: idempotency,
		now: func() time.Time {
			return time.Now().UTC()
		},
		newID: randomID,
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithClock(now func() time.Time) CheckoutOption {
	return func(service *CheckoutService) {
		service.now = now
	}
}

func WithIDGenerator(newID func(prefix string) string) CheckoutOption {
	return func(service *CheckoutService) {
		service.newID = newID
	}
}

func WithProductRepository(products ProductRepository) CheckoutOption {
	return func(service *CheckoutService) {
		service.products = products
	}
}

func WithStoreOperations(shifts ShiftRepository, days OperationalDayRepository) CheckoutOption {
	return func(service *CheckoutService) {
		service.shifts = shifts
		service.days = days
	}
}

type OpenReceiptCommand struct {
	IdempotencyKey string
	StoreID        string
	TerminalID     string
	CashierID      string
	Channel        string
}

type AddReceiptLineCommand struct {
	IdempotencyKey string
	ReceiptID      string
	ProductID      string
	Barcode        string
	Name           string
	Quantity       int64
	UnitPriceMinor int64
}

type ScanReceiptLineCommand struct {
	IdempotencyKey string
	ReceiptID      string
	Barcode        string
	Quantity       int64
}

type CancelReceiptCommand struct {
	IdempotencyKey string
	ReceiptID      string
	Reason         string
	ActorID        string
	ApprovedByID   string
}

type ReceiptResult struct {
	Receipt domain.Receipt
}

func (s *CheckoutService) OpenReceipt(ctx context.Context, command OpenReceiptCommand) (ReceiptResult, error) {
	if command.IdempotencyKey == "" {
		return ReceiptResult{}, ErrIdempotencyKeyRequired
	}
	if command.StoreID == "" || command.TerminalID == "" || command.CashierID == "" {
		return ReceiptResult{}, ErrInvalidCheckoutCommand
	}

	const operation = "checkout.open_receipt"
	channel := command.Channel
	if channel == "" {
		channel = "pos"
	}
	fingerprint := fmt.Sprintf("%s|%s|%s|%s", command.StoreID, command.TerminalID, command.CashierID, channel)
	if result, found, err := s.findReceiptIdempotency(ctx, operation, command.IdempotencyKey, "", fingerprint); err != nil || found {
		return result, err
	}

	var shift domain.Shift
	if s.shifts != nil {
		var err error
		shift, err = s.shifts.FindOpenShiftByTerminal(ctx, command.TerminalID)
		if err != nil {
			if errors.Is(err, ErrShiftNotFound) {
				return ReceiptResult{}, ErrOpenShiftRequired
			}
			return ReceiptResult{}, err
		}
		if shift.StoreID != command.StoreID || shift.CashierID != command.CashierID {
			return ReceiptResult{}, ErrOpenShiftRequired
		}
	}

	var day domain.OperationalDay
	if s.days != nil {
		var err error
		day, err = s.days.FindOpenOperationalDayByStore(ctx, command.StoreID)
		if err != nil {
			if errors.Is(err, ErrOperationalDayNotFound) {
				return ReceiptResult{}, ErrOpenOperationalDayRequired
			}
			return ReceiptResult{}, err
		}
	}

	receipt, err := domain.NewReceipt(domain.NewReceiptInput{
		ID:               s.newID("rct"),
		StoreID:          command.StoreID,
		OperationalDayID: day.ID,
		BusinessDate:     day.BusinessDate,
		ShiftID:          shift.ID,
		TerminalID:       command.TerminalID,
		CashierID:        command.CashierID,
		DrawerID:         shift.DrawerID,
		Channel:          channel,
		Now:              s.now(),
	})
	if err != nil {
		return ReceiptResult{}, err
	}

	if err := s.receipts.SaveReceipt(ctx, receipt); err != nil {
		return ReceiptResult{}, err
	}

	result := ReceiptResult{Receipt: receipt}
	if err := s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         command.IdempotencyKey,
		TargetID:    receipt.ID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now(),
	}); err != nil {
		return ReceiptResult{}, err
	}

	return result, nil
}

func (s *CheckoutService) GetReceipt(ctx context.Context, receiptID string) (ReceiptResult, error) {
	receipt, err := s.receipts.FindReceipt(ctx, receiptID)
	if err != nil {
		return ReceiptResult{}, err
	}
	return ReceiptResult{Receipt: receipt}, nil
}

func (s *CheckoutService) ListReceiptsByShift(ctx context.Context, shiftID string) ([]domain.Receipt, error) {
	if shiftID == "" {
		return nil, ErrInvalidCheckoutCommand
	}
	return s.receipts.ListReceiptsByShift(ctx, shiftID)
}

func (s *CheckoutService) ListReceiptsByOperationalDay(ctx context.Context, operationalDayID string) ([]domain.Receipt, error) {
	if operationalDayID == "" {
		return nil, ErrInvalidCheckoutCommand
	}
	return s.receipts.ListReceiptsByOperationalDay(ctx, operationalDayID)
}

func (s *CheckoutService) AddReceiptLine(ctx context.Context, command AddReceiptLineCommand) (ReceiptResult, error) {
	if command.IdempotencyKey == "" {
		return ReceiptResult{}, ErrIdempotencyKeyRequired
	}
	if command.ReceiptID == "" || command.ProductID == "" || command.Name == "" || command.Quantity <= 0 || command.UnitPriceMinor < 0 {
		return ReceiptResult{}, ErrInvalidCheckoutCommand
	}

	const operation = "checkout.add_receipt_line"
	fingerprint := fmt.Sprintf("%s|%s|%s|%s|%d|%d", command.ReceiptID, command.ProductID, command.Barcode, command.Name, command.Quantity, command.UnitPriceMinor)
	if result, found, err := s.findReceiptIdempotency(ctx, operation, command.IdempotencyKey, command.ReceiptID, fingerprint); err != nil || found {
		return result, err
	}

	receipt, err := s.receipts.FindReceipt(ctx, command.ReceiptID)
	if err != nil {
		return ReceiptResult{}, err
	}

	if err := receipt.AddLine(domain.AddReceiptLineInput{
		ID:             s.newID("rln"),
		ProductID:      command.ProductID,
		Barcode:        command.Barcode,
		Name:           command.Name,
		Quantity:       command.Quantity,
		UnitPriceMinor: command.UnitPriceMinor,
		Now:            s.now(),
	}); err != nil {
		return ReceiptResult{}, err
	}

	if err := s.receipts.SaveReceipt(ctx, receipt); err != nil {
		return ReceiptResult{}, err
	}

	result := ReceiptResult{Receipt: receipt}
	if err := s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         command.IdempotencyKey,
		TargetID:    command.ReceiptID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now(),
	}); err != nil {
		return ReceiptResult{}, err
	}

	return result, nil
}

func (s *CheckoutService) ScanReceiptLine(ctx context.Context, command ScanReceiptLineCommand) (ReceiptResult, error) {
	if s.products == nil {
		return ReceiptResult{}, ErrProductNotFound
	}
	if command.IdempotencyKey == "" {
		return ReceiptResult{}, ErrIdempotencyKeyRequired
	}
	if command.ReceiptID == "" || command.Barcode == "" || command.Quantity <= 0 {
		return ReceiptResult{}, ErrInvalidCheckoutCommand
	}

	const operation = "checkout.scan_receipt_line"
	fingerprint := fmt.Sprintf("%s|%s|%d", command.ReceiptID, command.Barcode, command.Quantity)
	if result, found, err := s.findReceiptIdempotency(ctx, operation, command.IdempotencyKey, command.ReceiptID, fingerprint); err != nil || found {
		return result, err
	}

	product, err := s.products.FindProductByBarcode(ctx, command.Barcode)
	if err != nil {
		return ReceiptResult{}, err
	}

	receipt, err := s.receipts.FindReceipt(ctx, command.ReceiptID)
	if err != nil {
		return ReceiptResult{}, err
	}

	if err := receipt.AddLine(domain.AddReceiptLineInput{
		ID:             s.newID("rln"),
		ProductID:      product.ID,
		Barcode:        command.Barcode,
		Name:           product.Name,
		Quantity:       command.Quantity,
		UnitPriceMinor: product.UnitPriceMinor,
		Now:            s.now(),
	}); err != nil {
		return ReceiptResult{}, err
	}

	if err := s.receipts.SaveReceipt(ctx, receipt); err != nil {
		return ReceiptResult{}, err
	}

	result := ReceiptResult{Receipt: receipt}
	if err := s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         command.IdempotencyKey,
		TargetID:    command.ReceiptID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now(),
	}); err != nil {
		return ReceiptResult{}, err
	}

	return result, nil
}

func (s *CheckoutService) CancelReceipt(ctx context.Context, command CancelReceiptCommand) (ReceiptResult, error) {
	if command.IdempotencyKey == "" {
		return ReceiptResult{}, ErrIdempotencyKeyRequired
	}
	if command.ReceiptID == "" || command.Reason == "" || command.ActorID == "" {
		return ReceiptResult{}, ErrInvalidCheckoutCommand
	}

	const operation = "checkout.cancel_receipt"
	fingerprint := fmt.Sprintf("%s|%s|%s|%s", command.ReceiptID, command.Reason, command.ActorID, command.ApprovedByID)
	if result, found, err := s.findReceiptIdempotency(ctx, operation, command.IdempotencyKey, command.ReceiptID, fingerprint); err != nil || found {
		return result, err
	}

	receipt, err := s.receipts.FindReceipt(ctx, command.ReceiptID)
	if err != nil {
		return ReceiptResult{}, err
	}
	if err := receipt.Cancel(domain.CancelReceiptInput{
		Reason:       command.Reason,
		ActorID:      command.ActorID,
		ApprovedByID: command.ApprovedByID,
		Now:          s.now(),
	}); err != nil {
		if errors.Is(err, domain.ErrReceiptCannotBeCancelled) {
			return ReceiptResult{}, ErrReceiptCannotBeCancelled
		}
		return ReceiptResult{}, err
	}

	if err := s.receipts.SaveReceipt(ctx, receipt); err != nil {
		return ReceiptResult{}, err
	}

	result := ReceiptResult{Receipt: receipt}
	if err := s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         command.IdempotencyKey,
		TargetID:    command.ReceiptID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now(),
	}); err != nil {
		return ReceiptResult{}, err
	}

	return result, nil
}

func (s *CheckoutService) findReceiptIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string) (ReceiptResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, operation, key)
	if err != nil || !found {
		return ReceiptResult{}, found, err
	}
	if targetID != "" && record.TargetID != targetID {
		return ReceiptResult{}, true, ErrIdempotencyKeyReused
	}
	if record.Fingerprint != fingerprint {
		return ReceiptResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(ReceiptResult)
	if !ok {
		return ReceiptResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

func randomID(prefix string) string {
	var bytes [12]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(fmt.Sprintf("generate id: %v", err))
	}
	return prefix + "_" + hex.EncodeToString(bytes[:])
}
