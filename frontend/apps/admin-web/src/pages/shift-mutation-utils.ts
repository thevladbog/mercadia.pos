import {
  getGetOperationalDaySummaryQueryKey,
  getListCashBalancesQueryKey,
  getListOpenStoreShiftsQueryKey,
} from '@mercadia/api-clients-store-edge';
import type { QueryClient } from '@tanstack/react-query';

export function parseNonNegativeRublesToMinor(value: string): number | null {
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
  if (!Number.isSafeInteger(minor) || minor < 0) {
    return null;
  }
  return minor;
}

export async function invalidateShiftCloseQueries(
  queryClient: QueryClient,
  storeId: string,
  operationalDayId?: string,
): Promise<void> {
  const tasks = [
    queryClient.invalidateQueries({ queryKey: getListOpenStoreShiftsQueryKey(storeId) }),
    queryClient.invalidateQueries({ queryKey: getListCashBalancesQueryKey(storeId) }),
  ];
  if (operationalDayId) {
    tasks.push(
      queryClient.invalidateQueries({
        queryKey: getGetOperationalDaySummaryQueryKey(operationalDayId),
      }),
    );
  }
  await Promise.all(tasks);
}
