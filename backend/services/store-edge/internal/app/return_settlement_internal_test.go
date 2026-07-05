package app

import (
	"errors"
	"testing"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

// mustCapturedRefundAllocationPayment builds a captured payment whose
// RefundableAmountMinor() equals amountMinor, using only the exported
// domain constructor.
func mustCapturedRefundAllocationPayment(t *testing.T, id string, amountMinor int64, capturedAt time.Time) domain.Payment {
	t.Helper()
	payment, err := domain.CreateCapturedPayment(domain.CreateCapturedPaymentInput{
		ID:          id,
		ReceiptID:   "receipt-1",
		Method:      domain.PaymentMethodCash,
		AmountMinor: amountMinor,
		Now:         capturedAt,
	})
	if err != nil {
		t.Fatalf("create captured payment %s: %v", id, err)
	}
	return payment
}

func TestAllocateRefundAmountsDistributesRoundingRemainder(t *testing.T) {
	base := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	payments := []domain.Payment{
		mustCapturedRefundAllocationPayment(t, "pay-1", 33, base),
		mustCapturedRefundAllocationPayment(t, "pay-2", 33, base.Add(time.Minute)),
		mustCapturedRefundAllocationPayment(t, "pay-3", 34, base.Add(2*time.Minute)),
	}

	allocations, err := allocateRefundAmounts(payments, 99, 100)
	if err != nil {
		t.Fatalf("allocateRefundAmounts returned error: %v", err)
	}

	var sum int64
	for _, payment := range payments {
		amount, ok := allocations[payment.ID]
		if !ok {
			continue
		}
		if amount > payment.RefundableAmountMinor() {
			t.Fatalf("payment %s allocated %d exceeds refundable %d", payment.ID, amount, payment.RefundableAmountMinor())
		}
		sum += amount
	}
	if sum != 99 {
		t.Fatalf("allocated sum = %d, want 99 (allocations=%v)", sum, allocations)
	}
}

func TestAllocateRefundAmountsFullRefundMatchesEachRefundable(t *testing.T) {
	base := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	payments := []domain.Payment{
		mustCapturedRefundAllocationPayment(t, "pay-1", 33, base),
		mustCapturedRefundAllocationPayment(t, "pay-2", 33, base.Add(time.Minute)),
		mustCapturedRefundAllocationPayment(t, "pay-3", 34, base.Add(2*time.Minute)),
	}

	allocations, err := allocateRefundAmounts(payments, 100, 100)
	if err != nil {
		t.Fatalf("allocateRefundAmounts returned error: %v", err)
	}

	for _, payment := range payments {
		amount, ok := allocations[payment.ID]
		if !ok {
			t.Fatalf("payment %s missing from allocations %v", payment.ID, allocations)
		}
		if amount != payment.RefundableAmountMinor() {
			t.Fatalf("payment %s allocated %d, want %d", payment.ID, amount, payment.RefundableAmountMinor())
		}
	}
}

func TestAllocateRefundAmountsSinglePaymentGetsWholeTotal(t *testing.T) {
	base := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	payments := []domain.Payment{
		mustCapturedRefundAllocationPayment(t, "pay-1", 50, base),
	}

	allocations, err := allocateRefundAmounts(payments, 50, 50)
	if err != nil {
		t.Fatalf("allocateRefundAmounts returned error: %v", err)
	}
	if len(allocations) != 1 || allocations["pay-1"] != 50 {
		t.Fatalf("allocations = %v, want {pay-1: 50}", allocations)
	}
}

func TestAllocateRefundAmountsRejectsTotalExceedingRefundable(t *testing.T) {
	base := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	payments := []domain.Payment{
		mustCapturedRefundAllocationPayment(t, "pay-1", 33, base),
		mustCapturedRefundAllocationPayment(t, "pay-2", 33, base.Add(time.Minute)),
		mustCapturedRefundAllocationPayment(t, "pay-3", 34, base.Add(2*time.Minute)),
	}

	_, err := allocateRefundAmounts(payments, 101, 100)
	if !errors.Is(err, ErrReturnSettlementPaymentMismatch) {
		t.Fatalf("err = %v, want ErrReturnSettlementPaymentMismatch", err)
	}
}

// TestAllocateRefundAmountsAlwaysSumsToReturnTotal is a property-style sweep:
// for every combination of 2-4 payments with varying refundable amounts and
// every return total from 1 up to the sum of refundables, the allocation
// must always sum exactly to the return total and never exceed any single
// payment's refundable amount.
func TestAllocateRefundAmountsAlwaysSumsToReturnTotal(t *testing.T) {
	base := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	refundableCombos := [][]int64{
		{33, 33, 34},
		{1, 1, 1},
		{7, 11},
		{10, 20, 30, 40},
		{5, 5, 5, 5},
		{1, 2, 3, 4},
		{100, 1},
		{17, 19, 23, 29},
	}

	for _, refundables := range refundableCombos {
		var totalRefundable int64
		payments := make([]domain.Payment, 0, len(refundables))
		for i, amount := range refundables {
			id := "pay-" + string(rune('a'+i))
			payments = append(payments, mustCapturedRefundAllocationPayment(t, id, amount, base.Add(time.Duration(i)*time.Minute)))
			totalRefundable += amount
		}

		for returnTotal := int64(1); returnTotal <= totalRefundable; returnTotal++ {
			allocations, err := allocateRefundAmounts(payments, returnTotal, totalRefundable)
			if err != nil {
				t.Fatalf("refundables=%v returnTotal=%d: unexpected error: %v", refundables, returnTotal, err)
			}
			var sum int64
			for _, payment := range payments {
				amount := allocations[payment.ID]
				if amount < 0 {
					t.Fatalf("refundables=%v returnTotal=%d: negative allocation for %s: %d", refundables, returnTotal, payment.ID, amount)
				}
				if amount > payment.RefundableAmountMinor() {
					t.Fatalf("refundables=%v returnTotal=%d: payment %s allocated %d exceeds refundable %d", refundables, returnTotal, payment.ID, amount, payment.RefundableAmountMinor())
				}
				sum += amount
			}
			if sum != returnTotal {
				t.Fatalf("refundables=%v returnTotal=%d: allocated sum = %d, want %d (allocations=%v)", refundables, returnTotal, sum, returnTotal, allocations)
			}
		}
	}
}
