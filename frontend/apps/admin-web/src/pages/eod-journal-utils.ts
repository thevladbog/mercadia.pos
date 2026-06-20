import { safeMovementHref, safeRecountHref } from './store-routes.js';

export type JournalReferenceAction =
  | { kind: 'shiftModal'; shiftId: string }
  | { kind: 'receiptModal'; receiptId: string }
  | { kind: 'returnModal'; returnId: string }
  | { kind: 'navigate'; href: string }
  | { kind: 'none' };

export function resolveJournalReferenceAction(
  operationType: string,
  referenceId: string | undefined,
  storeId: string,
): JournalReferenceAction {
  if (!referenceId) {
    return { kind: 'none' };
  }

  switch (operationType) {
    case 'shift.opened':
    case 'shift.closed':
      return { kind: 'shiftModal', shiftId: referenceId };
    case 'discount.applied':
      return { kind: 'receiptModal', receiptId: referenceId };
    case 'return.completed':
    case 'return.settled':
      return { kind: 'returnModal', returnId: referenceId };
    case 'cash.recount.created':
    case 'cash.recount.resolved':
      return { kind: 'navigate', href: safeRecountHref(storeId, referenceId) };
    case 'cash.movement.created':
      return { kind: 'navigate', href: safeMovementHref(storeId, referenceId) };
    default:
      return { kind: 'none' };
  }
}

export function journalOperationTypeKey(operationType: string): string {
  return `eod.operationTypes.${operationType}`;
}

export function isJournalReferenceActionable(action: JournalReferenceAction): boolean {
  return action.kind !== 'none';
}
