import type {
  CreateReceiptPayment202Payment,
  ListReceiptPayments200PaymentsItem,
} from '@mercadia/api-clients-store-edge';
import type { LayoutGridSpec } from '@mercadia/ui';

export type ReceiptPayment = CreateReceiptPayment202Payment | ListReceiptPayments200PaymentsItem;

export function formatMinorAmount(amountMinor: number, language: string): string {
  const locale = language === 'en' ? 'en-US' : 'ru-RU';
  return new Intl.NumberFormat(locale, { style: 'currency', currency: 'RUB' }).format(
    amountMinor / 100,
  );
}

export function filterGridByCategory(
  grid: LayoutGridSpec,
  categoryId: string | null,
): LayoutGridSpec {
  if (categoryId === null) {
    return grid;
  }
  return {
    ...grid,
    tiles: grid.tiles.filter((tile) => tile.categoryId === categoryId),
  };
}

export function parseAmountToMinor(value: string): number | null {
  const normalized = value.trim().replace(',', '.');
  if (!normalized) {
    return null;
  }
  if (!/^\d+(?:\.\d{0,2})?$/.test(normalized)) {
    return null;
  }
  const parsed = Number(normalized);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return null;
  }
  const [major, fractional = ''] = normalized.split('.');
  return Number(major) * 100 + Number(fractional.padEnd(2, '0'));
}

export function formatInputAmount(amountMinor: number): string {
  return (amountMinor / 100).toFixed(2);
}

export function settledPaymentAmountMinor(payment: ReceiptPayment): number {
  switch (payment.status) {
    case 'captured':
      return payment.amountMinor;
    case 'partially_refunded':
      return Math.max(payment.amountMinor - payment.refundedAmountMinor, 0);
    default:
      return 0;
  }
}
