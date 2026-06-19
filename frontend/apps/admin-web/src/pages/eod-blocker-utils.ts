import { safeRecountHref, storePageHref } from './store-routes.js';

export type EodTab = 'open-shifts';

export type BlockerAction =
  | { kind: 'eodTab'; tab: EodTab }
  | { kind: 'navigate'; href: string }
  | { kind: 'receiptModal'; receiptId: string }
  | { kind: 'shiftModal'; shiftId: string }
  | { kind: 'none' };

export function resolveBlockerAction(
  code: string,
  referenceId: string | undefined,
  storeId: string,
): BlockerAction {
  switch (code) {
    case 'open_cashier_shift':
      if (referenceId) {
        return { kind: 'shiftModal', shiftId: referenceId };
      }
      return { kind: 'eodTab', tab: 'open-shifts' };
    case 'unresolved_cash_recount_discrepancy':
      if (!referenceId) {
        return { kind: 'none' };
      }
      return { kind: 'navigate', href: safeRecountHref(storeId, referenceId) };
    case 'nonzero_drawer_balance':
      return { kind: 'navigate', href: storePageHref('/store/safe', storeId) };
    case 'unresolved_receipt':
      if (!referenceId) {
        return { kind: 'none' };
      }
      return { kind: 'receiptModal', receiptId: referenceId };
    case 'no_sales_receipts':
      return { kind: 'none' };
    default:
      return { kind: 'none' };
  }
}

export function resolveBlockerReferenceAction(
  code: string,
  referenceId: string | undefined,
  storeId: string,
): BlockerAction {
  if (code === 'open_cashier_shift' && referenceId) {
    return { kind: 'shiftModal', shiftId: referenceId };
  }
  return resolveBlockerAction(code, referenceId, storeId);
}

export function resolveBlockerHintAction(
  code: string,
  referenceId: string | undefined,
  storeId: string,
): BlockerAction {
  if (code === 'open_cashier_shift') {
    return { kind: 'eodTab', tab: 'open-shifts' };
  }
  return resolveBlockerAction(code, referenceId, storeId);
}

export function blockerActionHintKey(code: string): string | null {
  switch (code) {
    case 'open_cashier_shift':
      return 'eod.blockerActions.viewShift';
    case 'unresolved_cash_recount_discrepancy':
      return 'eod.blockerActions.resolveRecount';
    case 'nonzero_drawer_balance':
      return 'eod.blockerActions.viewSafe';
    case 'unresolved_receipt':
      return 'eod.blockerActions.viewReceipt';
    case 'no_sales_receipts':
      return 'eod.blockerActions.overrideInCloseModal';
    default:
      return null;
  }
}

export function isBlockerActionable(action: BlockerAction, code: string): boolean {
  if (action.kind !== 'none') {
    return true;
  }
  return code === 'no_sales_receipts';
}
