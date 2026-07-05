import {
  getGetCurrentOperationalDayQueryKey,
  getGetOperationalDaySummaryQueryKey,
} from '@mercadia/api-clients-store-edge';
import { QueryClient } from '@tanstack/react-query';
import type { TFunction } from 'i18next';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import {
  analyzeCloseReadiness,
  formatBlockerMessage,
  formatBlockerSeverity,
  invalidateEodAfterOpen,
  invalidateEodQueries,
  todayBusinessDate,
  type OperationalDayBlocker,
} from './eod-mutation-utils.js';

function blocker(overrides: Partial<OperationalDayBlocker>): OperationalDayBlocker {
  return {
    code: 'unresolved_shift',
    message: 'A shift is still open.',
    severity: 'blocker',
    ...overrides,
  };
}

describe('analyzeCloseReadiness', () => {
  it('allows a direct close when there are no blockers', () => {
    expect(analyzeCloseReadiness([])).toEqual({
      canCloseDirectly: true,
      canCloseWithOverride: false,
      isBlocked: false,
    });
  });

  it('allows an override close when only requires_admin_override blockers exist', () => {
    const blockers = [blocker({ severity: 'requires_admin_override' })];
    expect(analyzeCloseReadiness(blockers)).toEqual({
      canCloseDirectly: false,
      canCloseWithOverride: true,
      isBlocked: false,
    });
  });

  it('is blocked when a hard blocker is present', () => {
    const blockers = [blocker({ severity: 'blocker' })];
    expect(analyzeCloseReadiness(blockers)).toEqual({
      canCloseDirectly: false,
      canCloseWithOverride: false,
      isBlocked: true,
    });
  });

  it('is blocked (not override-closeable) when both a hard blocker and an override blocker exist', () => {
    const blockers = [
      blocker({ severity: 'blocker' }),
      blocker({ severity: 'requires_admin_override' }),
    ];
    expect(analyzeCloseReadiness(blockers)).toEqual({
      canCloseDirectly: false,
      canCloseWithOverride: false,
      isBlocked: true,
    });
  });
});

describe('invalidateEodQueries', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = new QueryClient();
  });

  it('invalidates the current operational day and the day summary', async () => {
    const spy = vi.spyOn(queryClient, 'invalidateQueries');

    await invalidateEodQueries(queryClient, 'store-1', 'day-1');

    expect(spy).toHaveBeenCalledTimes(2);
    expect(spy).toHaveBeenCalledWith({
      queryKey: getGetCurrentOperationalDayQueryKey('store-1'),
    });
    expect(spy).toHaveBeenCalledWith({
      queryKey: getGetOperationalDaySummaryQueryKey('day-1'),
    });
  });
});

describe('invalidateEodAfterOpen', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = new QueryClient();
  });

  it('invalidates only the current operational day when no operationalDayId is given', async () => {
    const spy = vi.spyOn(queryClient, 'invalidateQueries');

    await invalidateEodAfterOpen(queryClient, 'store-1');

    expect(spy).toHaveBeenCalledTimes(1);
    expect(spy).toHaveBeenCalledWith({
      queryKey: getGetCurrentOperationalDayQueryKey('store-1'),
    });
  });

  it('also invalidates the day summary when an operationalDayId is given', async () => {
    const spy = vi.spyOn(queryClient, 'invalidateQueries');

    await invalidateEodAfterOpen(queryClient, 'store-1', 'day-1');

    expect(spy).toHaveBeenCalledTimes(2);
    expect(spy).toHaveBeenCalledWith({
      queryKey: getGetOperationalDaySummaryQueryKey('day-1'),
    });
  });
});

describe('todayBusinessDate', () => {
  it('returns an ISO calendar date (YYYY-MM-DD)', () => {
    expect(todayBusinessDate()).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });
});

describe('formatBlockerSeverity / formatBlockerMessage', () => {
  const identityT = ((key: string) => key) as TFunction;

  it('falls back to the raw severity when no translation exists', () => {
    expect(formatBlockerSeverity('blocker', identityT)).toBe('blocker');
  });

  it('falls back to the blocker message when no translation exists for the code', () => {
    const b = blocker({ code: 'unresolved_shift', message: 'A shift is still open.' });
    expect(formatBlockerMessage(b, identityT)).toBe('A shift is still open.');
  });

  it('uses the translated value when it differs from the lookup key', () => {
    const translatingT = ((key: string) =>
      key === 'eod.severityLabels.blocker' ? 'Blocking issue' : key) as TFunction;
    expect(formatBlockerSeverity('blocker', translatingT)).toBe('Blocking issue');
  });
});
