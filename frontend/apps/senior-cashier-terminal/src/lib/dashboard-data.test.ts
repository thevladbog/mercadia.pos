import { describe, expect, it } from 'vitest';

import {
  deriveAlertsCount,
  deriveOperationsCount,
  deriveSafeFreshnessLabel,
  deriveTotalTerminalCount,
  joinCashiersOnShift,
} from './dashboard-data.js';

describe('joinCashiersOnShift', () => {
  const shift = {
    id: 'shift-1',
    cashierId: 'cashier-1',
    terminalId: 'term-1',
    openedAt: '2026-07-09T08:00:00.000Z',
  };

  it('joins a shift with its matching monitoring terminal and credential actor', () => {
    const result = joinCashiersOnShift(
      [shift],
      [{ id: 'term-1', revenueMinor: 184_820_00, drawerBalanceMinor: 96_420_00 }],
      [{ id: 'cashier-1', roles: ['cashier'] }],
    );
    expect(result).toEqual([
      {
        shiftId: 'shift-1',
        cashierId: 'cashier-1',
        role: 'cashier',
        terminalId: 'term-1',
        openedAt: '2026-07-09T08:00:00.000Z',
        revenueMinor: 184_820_00,
        drawerBalanceMinor: 96_420_00,
      },
    ]);
  });

  it('defaults revenue/drawer balance to 0 when no monitoring terminal matches, without throwing', () => {
    const result = joinCashiersOnShift([shift], [], [{ id: 'cashier-1', roles: ['cashier'] }]);
    expect(result[0].revenueMinor).toBe(0);
    expect(result[0].drawerBalanceMinor).toBe(0);
  });

  it('sets role to null when no credential actor matches', () => {
    const result = joinCashiersOnShift(
      [shift],
      [{ id: 'term-1', revenueMinor: 0, drawerBalanceMinor: 0 }],
      [],
    );
    expect(result[0].role).toBeNull();
  });

  it('sets role to null when the matching actor has an empty roles array', () => {
    const result = joinCashiersOnShift([shift], [], [{ id: 'cashier-1', roles: [] }]);
    expect(result[0].role).toBeNull();
  });

  it('returns an empty array for no open shifts', () => {
    expect(joinCashiersOnShift([], [], [])).toEqual([]);
  });

  it('joins multiple shifts independently by their own terminalId/cashierId', () => {
    const shifts = [
      shift,
      {
        id: 'shift-2',
        cashierId: 'cashier-2',
        terminalId: 'term-2',
        openedAt: '2026-07-09T09:00:00.000Z',
      },
    ];
    const terminals = [
      { id: 'term-1', revenueMinor: 100, drawerBalanceMinor: 10 },
      { id: 'term-2', revenueMinor: 200, drawerBalanceMinor: 20 },
    ];
    const actors = [
      { id: 'cashier-1', roles: ['cashier'] },
      { id: 'cashier-2', roles: ['senior_cashier'] },
    ];
    const result = joinCashiersOnShift(shifts, terminals, actors);
    expect(result).toEqual([
      {
        shiftId: 'shift-1',
        cashierId: 'cashier-1',
        role: 'cashier',
        terminalId: 'term-1',
        openedAt: '2026-07-09T08:00:00.000Z',
        revenueMinor: 100,
        drawerBalanceMinor: 10,
      },
      {
        shiftId: 'shift-2',
        cashierId: 'cashier-2',
        role: 'senior_cashier',
        terminalId: 'term-2',
        openedAt: '2026-07-09T09:00:00.000Z',
        revenueMinor: 200,
        drawerBalanceMinor: 20,
      },
    ]);
  });
});

describe('deriveOperationsCount', () => {
  const actorId = 'senior-1';
  const sinceIso = '2026-07-09T10:00:00.000Z';

  it('counts only entries authored by actorId at or after sinceIso', () => {
    const entries = [
      { actorId: 'senior-1', createdAt: '2026-07-09T10:00:00.000Z' },
      { actorId: 'senior-1', createdAt: '2026-07-09T11:00:00.000Z' },
      { actorId: 'senior-1', createdAt: '2026-07-09T09:59:59.000Z' },
      { actorId: 'cashier-1', createdAt: '2026-07-09T12:00:00.000Z' },
    ];
    expect(deriveOperationsCount(entries, actorId, sinceIso)).toBe(2);
  });

  it('returns 0 for an empty journal', () => {
    expect(deriveOperationsCount([], actorId, sinceIso)).toBe(0);
  });

  it('returns 0 when no entries match the actor', () => {
    const entries = [{ actorId: 'someone-else', createdAt: '2026-07-09T12:00:00.000Z' }];
    expect(deriveOperationsCount(entries, actorId, sinceIso)).toBe(0);
  });
});

describe('deriveAlertsCount', () => {
  it('sums attentionTerminalCount and offlineTerminalCount', () => {
    expect(deriveAlertsCount({ attentionTerminalCount: 2, offlineTerminalCount: 3 })).toBe(5);
  });

  it('returns 0 for an undefined summary', () => {
    expect(deriveAlertsCount(undefined)).toBe(0);
  });

  it('returns 0 when both counts are 0', () => {
    expect(deriveAlertsCount({ attentionTerminalCount: 0, offlineTerminalCount: 0 })).toBe(0);
  });
});

describe('deriveTotalTerminalCount', () => {
  it('sums active, free, and offline terminal counts', () => {
    expect(
      deriveTotalTerminalCount({
        activeTerminalCount: 5,
        freeTerminalCount: 2,
        offlineTerminalCount: 1,
      }),
    ).toBe(8);
  });

  it('returns undefined (not 0) when the summary has not loaded yet', () => {
    expect(deriveTotalTerminalCount(undefined)).toBeUndefined();
  });

  it('returns 0 when the summary loaded with all-zero counts', () => {
    expect(
      deriveTotalTerminalCount({
        activeTerminalCount: 0,
        freeTerminalCount: 0,
        offlineTerminalCount: 0,
      }),
    ).toBe(0);
  });
});

describe('deriveSafeFreshnessLabel', () => {
  const translate = (key: string, params?: Record<string, unknown>) =>
    params ? `[${key}:${JSON.stringify(params)}]` : `[${key}]`;
  const now = new Date(2026, 6, 9, 14, 30, 0).getTime();

  it('formats a sub-minute difference in seconds', () => {
    const at = new Date(now - 5_000).toISOString();
    expect(deriveSafeFreshnessLabel(at, now, translate)).toBe(
      '[dashboard.safeNow.freshnessSeconds:{"count":5}]',
    );
  });

  it('formats a sub-hour difference in minutes', () => {
    const at = new Date(now - 5 * 60_000).toISOString();
    expect(deriveSafeFreshnessLabel(at, now, translate)).toBe(
      '[dashboard.safeNow.freshnessMinutes:{"count":5}]',
    );
  });

  it('formats an hour-or-more difference in hours', () => {
    const at = new Date(now - 3 * 60 * 60_000).toISOString();
    expect(deriveSafeFreshnessLabel(at, now, translate)).toBe(
      '[dashboard.safeNow.freshnessHours:{"count":3}]',
    );
  });

  it('falls back to the unknown label for a missing timestamp, without throwing', () => {
    expect(deriveSafeFreshnessLabel(undefined, now, translate)).toBe(
      '[dashboard.safeNow.freshnessUnknown]',
    );
  });

  it('falls back to the unknown label for an unparseable timestamp, without throwing', () => {
    expect(deriveSafeFreshnessLabel('not-a-date', now, translate)).toBe(
      '[dashboard.safeNow.freshnessUnknown]',
    );
  });
});
