import {
  getListCashBalancesQueryKey,
  getListCashMovementsQueryKey,
  getListCashRecountsQueryKey,
} from '@mercadia/api-clients-store-edge';
import { QueryClient } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import {
  actorsMustDiffer,
  formatMinorToRublesInput,
  invalidateSafeQueries,
  parseRublesToMinor,
} from './cash-mutation-utils.js';

describe('parseRublesToMinor', () => {
  it('parses a plain decimal amount to minor units', () => {
    expect(parseRublesToMinor('100.50')).toBe(10050);
  });

  it('accepts a comma decimal separator', () => {
    expect(parseRublesToMinor('12,5')).toBe(1250);
  });

  it('rejects a zero amount', () => {
    expect(parseRublesToMinor('0')).toBeNull();
  });

  it('rejects an amount with more than two fractional digits', () => {
    expect(parseRublesToMinor('100.123')).toBeNull();
  });

  it('rejects an empty string', () => {
    expect(parseRublesToMinor('   ')).toBeNull();
  });
});

describe('formatMinorToRublesInput / parseRublesToMinor round-trip', () => {
  it('round-trips minor units back through the rubles formatter', () => {
    const minor = parseRublesToMinor('100.50');
    expect(minor).not.toBeNull();
    expect(formatMinorToRublesInput(minor as number)).toBe('100.50');
  });
});

describe('actorsMustDiffer', () => {
  it('returns false when the actor and approver ids are the same', () => {
    expect(actorsMustDiffer('staff-1', 'staff-1')).toBe(false);
  });

  it('returns true when the actor and approver ids differ', () => {
    expect(actorsMustDiffer('staff-1', 'staff-2')).toBe(true);
  });

  it('characterization: returns false when either id is empty, even if they differ', () => {
    expect(actorsMustDiffer('', 'staff-2')).toBe(false);
    expect(actorsMustDiffer('staff-1', '')).toBe(false);
    expect(actorsMustDiffer('', '')).toBe(false);
  });
});

describe('invalidateSafeQueries', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = new QueryClient();
  });

  it('invalidates cash balances, movements, and recounts for the store', async () => {
    const spy = vi.spyOn(queryClient, 'invalidateQueries');

    await invalidateSafeQueries(queryClient, 'store-1');

    expect(spy).toHaveBeenCalledTimes(3);
    expect(spy).toHaveBeenCalledWith({ queryKey: getListCashBalancesQueryKey('store-1') });
    expect(spy).toHaveBeenCalledWith({ queryKey: getListCashMovementsQueryKey('store-1') });
    expect(spy).toHaveBeenCalledWith({ queryKey: getListCashRecountsQueryKey('store-1') });
  });
});
