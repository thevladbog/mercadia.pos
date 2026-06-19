import type { QueryClient } from '@tanstack/react-query';
import {
  getListCashBalancesQueryKey,
  getListCashMovementsQueryKey,
  getListCashRecountsQueryKey,
} from '@mercadia/api-clients-store-edge';

export function createIdempotencyHeaders(): HeadersInit {
  return { 'Idempotency-Key': crypto.randomUUID() };
}

export function parseRublesToMinor(value: string): number | null {
  const normalized = value.trim().replace(',', '.');
  if (normalized.length === 0) {
    return null;
  }
  const match = /^(\d+)(?:\.(\d{0,2}))?$/.exec(normalized);
  if (!match) {
    return null;
  }
  const fractionalPart = (match[2] ?? '').padEnd(2, '0');
  const minor = Number(match[1]) * 100 + Number(fractionalPart);
  if (!Number.isSafeInteger(minor) || minor <= 0) {
    return null;
  }
  return minor;
}

export function formatMinorToRublesInput(minor: number): string {
  return (minor / 100).toFixed(2);
}

export async function invalidateSafeQueries(
  queryClient: QueryClient,
  storeId: string,
): Promise<void> {
  await Promise.all([
    queryClient.invalidateQueries({ queryKey: getListCashBalancesQueryKey(storeId) }),
    queryClient.invalidateQueries({ queryKey: getListCashMovementsQueryKey(storeId) }),
    queryClient.invalidateQueries({ queryKey: getListCashRecountsQueryKey(storeId) }),
  ]);
}

export function actorsMustDiffer(actorId: string, approvedById: string): boolean {
  const actor = actorId.trim();
  const approver = approvedById.trim();
  return actor.length > 0 && approver.length > 0 && actor !== approver;
}
