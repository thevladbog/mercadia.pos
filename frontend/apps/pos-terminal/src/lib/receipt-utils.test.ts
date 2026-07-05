import type {
  CreateReceiptPayment202Payment,
  ListReceiptPayments200PaymentsItem,
} from '@mercadia/api-clients-store-edge';
import type { LayoutGridSpec } from '@mercadia/ui';
import { describe, expect, it } from 'vitest';

import {
  filterGridByCategory,
  formatInputAmount,
  formatMinorAmount,
  parseAmountToMinor,
  settledPaymentAmountMinor,
} from './receipt-utils.js';

function buildGrid(): LayoutGridSpec {
  return {
    rows: 2,
    cols: 2,
    categories: [
      { id: 'drinks', label: 'Drinks' },
      { id: 'snacks', label: 'Snacks' },
    ],
    tiles: [
      { label: 'Cola', categoryId: 'drinks' },
      { label: 'Chips', categoryId: 'snacks' },
      { label: 'Water', categoryId: 'drinks' },
    ],
  };
}

function buildPayment(
  overrides: Partial<ListReceiptPayments200PaymentsItem>,
): CreateReceiptPayment202Payment | ListReceiptPayments200PaymentsItem {
  return {
    amountMinor: 10000,
    capturedAt: '2026-07-05T10:00:00.000Z',
    createdAt: '2026-07-05T10:00:00.000Z',
    id: 'payment-1',
    method: 'cash',
    receiptId: 'receipt-1',
    refundedAmountMinor: 0,
    status: 'captured',
    updatedAt: '2026-07-05T10:00:00.000Z',
    ...overrides,
  };
}

describe('filterGridByCategory', () => {
  it('returns the full grid when categoryId is null', () => {
    const grid = buildGrid();
    expect(filterGridByCategory(grid, null)).toEqual(grid);
  });

  it('filters tiles down to the matching category', () => {
    const grid = buildGrid();
    const filtered = filterGridByCategory(grid, 'drinks');
    expect(filtered.tiles).toEqual([
      { label: 'Cola', categoryId: 'drinks' },
      { label: 'Water', categoryId: 'drinks' },
    ]);
    expect(filtered.rows).toBe(grid.rows);
    expect(filtered.cols).toBe(grid.cols);
  });

  it('returns an empty tiles array for a category with no matching tiles', () => {
    const grid = buildGrid();
    expect(filterGridByCategory(grid, 'unknown-category').tiles).toEqual([]);
  });
});

describe('settledPaymentAmountMinor', () => {
  it('returns the full amount for a captured payment', () => {
    expect(settledPaymentAmountMinor(buildPayment({ status: 'captured', amountMinor: 5000 }))).toBe(
      5000,
    );
  });

  it('returns amount minus refunded for a partially refunded payment', () => {
    expect(
      settledPaymentAmountMinor(
        buildPayment({
          status: 'partially_refunded',
          amountMinor: 5000,
          refundedAmountMinor: 2000,
        }),
      ),
    ).toBe(3000);
  });

  it('clamps a partially refunded payment at 0 when refunded exceeds amount', () => {
    expect(
      settledPaymentAmountMinor(
        buildPayment({
          status: 'partially_refunded',
          amountMinor: 5000,
          refundedAmountMinor: 9000,
        }),
      ),
    ).toBe(0);
  });

  it('returns 0 for a cancelled payment', () => {
    expect(
      settledPaymentAmountMinor(buildPayment({ status: 'cancelled', amountMinor: 5000 })),
    ).toBe(0);
  });

  it('returns 0 for a fully refunded payment', () => {
    expect(settledPaymentAmountMinor(buildPayment({ status: 'refunded', amountMinor: 5000 }))).toBe(
      0,
    );
  });
});

describe('parseAmountToMinor / formatInputAmount round-trip', () => {
  it('parses a plain decimal amount to minor units', () => {
    expect(parseAmountToMinor('100.50')).toBe(10050);
  });

  it('round-trips minor units back through the input formatter', () => {
    const minor = parseAmountToMinor('100.50');
    expect(minor).not.toBeNull();
    expect(formatInputAmount(minor as number)).toBe('100.50');
  });

  it('accepts a comma decimal separator', () => {
    expect(parseAmountToMinor('12,5')).toBe(1250);
  });

  it('rejects a zero amount', () => {
    expect(parseAmountToMinor('0')).toBeNull();
  });

  it('rejects a negative amount', () => {
    expect(parseAmountToMinor('-5')).toBeNull();
  });

  it('rejects non-numeric input', () => {
    expect(parseAmountToMinor('abc')).toBeNull();
  });

  it('rejects an empty string', () => {
    expect(parseAmountToMinor('   ')).toBeNull();
  });
});

describe('formatMinorAmount', () => {
  it('formats as RUB currency for the ru language', () => {
    const formatted = formatMinorAmount(10050, 'ru');
    expect(formatted).toContain('100,50');
    expect(formatted).toContain('₽');
  });

  it('formats as RUB currency for the en language', () => {
    const formatted = formatMinorAmount(10050, 'en');
    expect(formatted).toContain('100.50');
    expect(formatted).toContain('RUB');
  });
});
