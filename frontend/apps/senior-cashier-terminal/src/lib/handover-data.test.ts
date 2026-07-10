import { describe, expect, it } from 'vitest';

import {
  deriveHandoverSummary,
  deriveHandoverWarnings,
  joinEligibleSuccessors,
} from './handover-data.js';

describe('joinEligibleSuccessors', () => {
  it('includes senior_cashier and admin actors, excluding the current session actor', () => {
    const result = joinEligibleSuccessors(
      [
        { id: 'senior-1', roles: ['senior_cashier'] },
        { id: 'senior-2', roles: ['senior_cashier'] },
        { id: 'admin-1', roles: ['admin'] },
        { id: 'cashier-1', roles: ['cashier'] },
      ],
      'senior-1',
    );
    expect(result).toEqual([
      { actorId: 'senior-2', role: 'senior_cashier' },
      { actorId: 'admin-1', role: 'admin' },
    ]);
  });

  it('excludes actors with no senior_cashier/admin role', () => {
    const result = joinEligibleSuccessors(
      [
        { id: 'cashier-1', roles: ['cashier'] },
        { id: 'assist-1', roles: ['assistant'] },
      ],
      'senior-1',
    );
    expect(result).toEqual([]);
  });

  it('returns null role for an actor with an empty roles array (defensive, should not normally occur post-filter)', () => {
    // roles=[] never passes the senior_cashier/admin filter, so this exercises
    // the ?? null fallback only in a hypothetical multi-role edge case.
    const result = joinEligibleSuccessors(
      [{ id: 'senior-2', roles: ['admin', 'senior_cashier'] }],
      'senior-1',
    );
    expect(result).toEqual([{ actorId: 'senior-2', role: 'admin' }]);
  });

  it('returns an empty array when the only actor is the excluded one', () => {
    expect(
      joinEligibleSuccessors([{ id: 'senior-1', roles: ['senior_cashier'] }], 'senior-1'),
    ).toEqual([]);
  });
});

describe('deriveHandoverSummary', () => {
  const actorId = 'senior-1';
  const otherActorId = 'senior-2';
  const sinceIso = '2026-07-10T08:00:00.000Z';
  const before = '2026-07-10T07:00:00.000Z';
  const after = '2026-07-10T09:00:00.000Z';

  it('sums change_fund movements where approvedById matches (3 auto-derived pages use approvedById)', () => {
    const movements = [
      {
        type: 'change_fund',
        fromContainerType: 'safe',
        toContainerType: 'drawer',
        amountMinor: 10_000,
        actorId: 'cashier-1',
        approvedById: actorId,
        createdAt: after,
      },
      {
        type: 'change_fund',
        fromContainerType: 'safe',
        toContainerType: 'drawer',
        amountMinor: 5_000,
        actorId: 'cashier-2',
        approvedById: otherActorId,
        createdAt: after,
      },
    ];
    const summary = deriveHandoverSummary(movements, [], actorId, sinceIso);
    expect(summary.changeIssuedMinor).toBe(10_000);
    expect(summary.changeIssuedCount).toBe(1);
  });

  it('sums cash_out movements where approvedById matches', () => {
    const movements = [
      {
        type: 'cash_out',
        fromContainerType: 'drawer',
        toContainerType: 'safe',
        amountMinor: 20_000,
        actorId: 'cashier-1',
        approvedById: actorId,
        createdAt: after,
      },
    ];
    const summary = deriveHandoverSummary(movements, [], actorId, sinceIso);
    expect(summary.cashReceivedMinor).toBe(20_000);
    expect(summary.cashReceivedCount).toBe(1);
  });

  it('sums drawer_to_safe movements where approvedById matches (final incassations)', () => {
    const movements = [
      {
        type: 'drawer_to_safe',
        fromContainerType: 'drawer',
        toContainerType: 'safe',
        amountMinor: 30_000,
        actorId: 'cashier-1',
        approvedById: actorId,
        createdAt: after,
      },
    ];
    const summary = deriveHandoverSummary(movements, [], actorId, sinceIso);
    expect(summary.incassationsMinor).toBe(30_000);
    expect(summary.incassationsCount).toBe(1);
  });

  it('sums expense movements via the actorId-OR-approvedById fallback (manual-entry page)', () => {
    // BusinessExpensePage keeps actorId/approvedById as manual free-text
    // inputs (plan 022) — this plan's OR-condition fallback must still
    // capture a movement where actorId (not approvedById) matches the
    // closing senior cashier.
    const movements = [
      {
        type: 'expense',
        fromContainerType: 'safe',
        toContainerType: 'external',
        amountMinor: 7_500,
        actorId,
        approvedById: otherActorId,
        createdAt: after,
      },
    ];
    const summary = deriveHandoverSummary(movements, [], actorId, sinceIso);
    expect(summary.expensesMinor).toBe(7_500);
    expect(summary.expensesCount).toBe(1);
  });

  it('excludes movements before sinceIso even if the actor filter matches', () => {
    const movements = [
      {
        type: 'change_fund',
        fromContainerType: 'safe',
        toContainerType: 'drawer',
        amountMinor: 10_000,
        actorId: 'cashier-1',
        approvedById: actorId,
        createdAt: before,
      },
    ];
    const summary = deriveHandoverSummary(movements, [], actorId, sinceIso);
    expect(summary.changeIssuedMinor).toBe(0);
    expect(summary.changeIssuedCount).toBe(0);
  });

  it('excludes movements neither actorId nor approvedById match', () => {
    const movements = [
      {
        type: 'change_fund',
        fromContainerType: 'safe',
        toContainerType: 'drawer',
        amountMinor: 10_000,
        actorId: 'cashier-1',
        approvedById: otherActorId,
        createdAt: after,
      },
    ];
    const summary = deriveHandoverSummary(movements, [], actorId, sinceIso);
    expect(summary.changeIssuedMinor).toBe(0);
  });

  it('computes safeNetChangeMinor as a signed sum across every safe-touching movement matching the actor filter', () => {
    const movements = [
      // safe -> drawer: subtracts from the safe.
      {
        type: 'change_fund',
        fromContainerType: 'safe',
        toContainerType: 'drawer',
        amountMinor: 10_000,
        actorId: 'cashier-1',
        approvedById: actorId,
        createdAt: after,
      },
      // drawer -> safe: adds to the safe.
      {
        type: 'drawer_to_safe',
        fromContainerType: 'drawer',
        toContainerType: 'safe',
        amountMinor: 30_000,
        actorId: 'cashier-1',
        approvedById: actorId,
        createdAt: after,
      },
      // safe -> bank (safe_to_bank has no row of its own but DOES count here).
      {
        type: 'safe_to_bank',
        fromContainerType: 'safe',
        toContainerType: 'bank',
        amountMinor: 4_000,
        actorId,
        approvedById: otherActorId,
        createdAt: after,
      },
    ];
    const summary = deriveHandoverSummary(movements, [], actorId, sinceIso);
    expect(summary.safeNetChangeMinor).toBe(30_000 - 10_000 - 4_000);
    expect(summary.safeNetChangeCount).toBe(3);
  });

  it('counts a return as approved only when approvedById matches, confirmedAt is real, and confirmedAt >= sinceIso', () => {
    const returns = [
      // Real confirmation, in-session: counts.
      { approvedById: actorId, confirmedAt: after, totalMinor: 1_500 },
      // approvedById does not match: excluded.
      { approvedById: otherActorId, confirmedAt: after, totalMinor: 999 },
      // Zero-value confirmedAt (Go zero time) — a with-receipt return
      // completed at creation, never routed through the confirm flow.
      { approvedById: actorId, confirmedAt: '0001-01-01T00:00:00Z', totalMinor: 999 },
      // No confirmedAt at all.
      { approvedById: actorId, totalMinor: 999 },
      // Confirmed, but before this session started.
      { approvedById: actorId, confirmedAt: before, totalMinor: 999 },
    ];
    const summary = deriveHandoverSummary([], returns, actorId, sinceIso);
    expect(summary.refundsApprovedMinor).toBe(1_500);
    expect(summary.refundsApprovedCount).toBe(1);
  });

  it('returns all zeros for empty inputs', () => {
    const summary = deriveHandoverSummary([], [], actorId, sinceIso);
    expect(summary).toEqual({
      changeIssuedMinor: 0,
      changeIssuedCount: 0,
      cashReceivedMinor: 0,
      cashReceivedCount: 0,
      incassationsMinor: 0,
      incassationsCount: 0,
      expensesMinor: 0,
      expensesCount: 0,
      refundsApprovedMinor: 0,
      refundsApprovedCount: 0,
      safeNetChangeMinor: 0,
      safeNetChangeCount: 0,
    });
  });
});

describe('deriveHandoverWarnings', () => {
  it('counts requested change-fund requests and open recount discrepancies', () => {
    const result = deriveHandoverWarnings(
      [{ status: 'requested' }, { status: 'requested' }, { status: 'fulfilled' }],
      [
        { resolutionStatus: 'open' },
        { resolutionStatus: 'resolved' },
        { resolutionStatus: 'not_required' },
      ],
    );
    expect(result).toEqual({ pendingChangeFundRequests: 2, openRecountDiscrepancies: 1 });
  });

  it('returns zero counts for empty inputs', () => {
    expect(deriveHandoverWarnings([], [])).toEqual({
      pendingChangeFundRequests: 0,
      openRecountDiscrepancies: 0,
    });
  });

  it('returns zero counts when nothing matches the pending/open statuses', () => {
    expect(
      deriveHandoverWarnings([{ status: 'fulfilled' }], [{ resolutionStatus: 'resolved' }]),
    ).toEqual({ pendingChangeFundRequests: 0, openRecountDiscrepancies: 0 });
  });
});
