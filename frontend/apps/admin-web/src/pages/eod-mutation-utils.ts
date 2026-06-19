import type { GetOperationalDaySummary200BlockersItem } from '@mercadia/api-clients-store-edge';
import {
  getGetCurrentOperationalDayQueryKey,
  getGetOperationalDaySummaryQueryKey,
} from '@mercadia/api-clients-store-edge';
import type { QueryClient } from '@tanstack/react-query';
import type { TFunction } from 'i18next';

export type OperationalDayBlocker = GetOperationalDaySummary200BlockersItem;

export type CloseReadiness = {
  canCloseDirectly: boolean;
  canCloseWithOverride: boolean;
  isBlocked: boolean;
};

export function analyzeCloseReadiness(blockers: OperationalDayBlocker[]): CloseReadiness {
  const hard = blockers.filter((blocker) => blocker.severity === 'blocker');
  const overrides = blockers.filter((blocker) => blocker.severity === 'requires_admin_override');

  return {
    canCloseDirectly: blockers.length === 0,
    canCloseWithOverride: hard.length === 0 && overrides.length > 0,
    isBlocked: hard.length > 0,
  };
}

export async function invalidateEodQueries(
  queryClient: QueryClient,
  storeId: string,
  operationalDayId: string,
): Promise<void> {
  await Promise.all([
    queryClient.invalidateQueries({ queryKey: getGetCurrentOperationalDayQueryKey(storeId) }),
    queryClient.invalidateQueries({
      queryKey: getGetOperationalDaySummaryQueryKey(operationalDayId),
    }),
  ]);
}

export function todayBusinessDate(): string {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, '0');
  const day = String(now.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

export async function invalidateEodAfterOpen(
  queryClient: QueryClient,
  storeId: string,
  operationalDayId?: string,
): Promise<void> {
  const tasks = [
    queryClient.invalidateQueries({ queryKey: getGetCurrentOperationalDayQueryKey(storeId) }),
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

export function formatBlockerSeverity(severity: string, t: TFunction): string {
  const key = `eod.severityLabels.${severity}`;
  const translated = t(key);
  return translated === key ? severity : translated;
}

export function formatBlockerMessage(blocker: OperationalDayBlocker, t: TFunction): string {
  const codeKey = `eod.blockerCodes.${blocker.code}`;
  const translated = t(codeKey);
  return translated === codeKey ? blocker.message : translated;
}
