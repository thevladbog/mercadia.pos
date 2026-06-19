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
  const rubles = Number(normalized);
  if (!Number.isFinite(rubles) || rubles <= 0) {
    return null;
  }
  return Math.round(rubles * 100);
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
  return actorId.trim().length > 0 && approvedById.trim().length > 0 && actorId !== approvedById;
}
