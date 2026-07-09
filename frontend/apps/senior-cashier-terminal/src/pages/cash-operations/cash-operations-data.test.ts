import { describe, expect, it } from 'vitest';

import {
  computeAfterBalance,
  findActorRole,
  findContainerBalance,
  findSafeBalance,
} from './cash-operations-data.js';

describe('findContainerBalance', () => {
  const balances = [
    { containerId: 'drawer-1', containerType: 'drawer', balanceMinor: 10_000 },
    { containerId: 'drawer-2', containerType: 'drawer', balanceMinor: 25_000 },
    { containerId: 'safe-1', containerType: 'safe', balanceMinor: 500_000 },
  ];

  it('finds the balance matching the exact containerId, not the first of that type (regression for plan 022 item 4)', () => {
    expect(findContainerBalance(balances, 'drawer-2')).toEqual({
      containerId: 'drawer-2',
      containerType: 'drawer',
      balanceMinor: 25_000,
    });
  });

  it('picks a different drawer correctly when multiple drawers are open', () => {
    expect(findContainerBalance(balances, 'drawer-1')).toEqual({
      containerId: 'drawer-1',
      containerType: 'drawer',
      balanceMinor: 10_000,
    });
  });

  it('returns undefined for an unknown containerId', () => {
    expect(findContainerBalance(balances, 'drawer-99')).toBeUndefined();
  });

  it('returns undefined when containerId is undefined', () => {
    expect(findContainerBalance(balances, undefined)).toBeUndefined();
  });

  it('returns undefined for an empty balances array', () => {
    expect(findContainerBalance([], 'drawer-1')).toBeUndefined();
  });
});

describe('findSafeBalance', () => {
  it('finds the balance with containerType "safe"', () => {
    const balances = [
      { containerId: 'drawer-1', containerType: 'drawer', balanceMinor: 10_000 },
      { containerId: 'safe-1', containerType: 'safe', balanceMinor: 500_000 },
    ];
    expect(findSafeBalance(balances)).toEqual({
      containerId: 'safe-1',
      containerType: 'safe',
      balanceMinor: 500_000,
    });
  });

  it('returns undefined when no safe container exists', () => {
    expect(
      findSafeBalance([{ containerId: 'drawer-1', containerType: 'drawer', balanceMinor: 1 }]),
    ).toBeUndefined();
  });

  it('returns undefined for an empty balances array', () => {
    expect(findSafeBalance([])).toBeUndefined();
  });
});

describe('computeAfterBalance', () => {
  it('adds the total for an "increase" direction (e.g. cash moving into the safe)', () => {
    expect(computeAfterBalance(100_000, 25_000, 'increase')).toBe(125_000);
  });

  it('subtracts the total for a "decrease" direction (e.g. cash moving out of the safe)', () => {
    expect(computeAfterBalance(100_000, 25_000, 'decrease')).toBe(75_000);
  });

  it('treats a zero total as a no-op', () => {
    expect(computeAfterBalance(100_000, 0, 'increase')).toBe(100_000);
  });

  it('allows the after balance to go negative (no clamping — a real discrepancy should be visible, not hidden)', () => {
    expect(computeAfterBalance(1_000, 5_000, 'decrease')).toBe(-4_000);
  });
});

describe('findActorRole', () => {
  const actors = [
    { id: 'cashier-1', roles: ['cashier'] },
    { id: 'cashier-2', roles: [] },
  ];

  it('returns the matching actor first role', () => {
    expect(findActorRole(actors, 'cashier-1')).toBe('cashier');
  });

  it('returns null when the matching actor has an empty roles array', () => {
    expect(findActorRole(actors, 'cashier-2')).toBeNull();
  });

  it('returns null when no actor matches', () => {
    expect(findActorRole(actors, 'unknown')).toBeNull();
  });

  it('returns null when actorId is undefined', () => {
    expect(findActorRole(actors, undefined)).toBeNull();
  });
});
