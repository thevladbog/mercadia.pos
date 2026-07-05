import { describe, expect, it } from 'vitest';

import { actorsMustDiffer, computeDenominationTotal, selectSuccessData } from './cash-utils.js';

describe('computeDenominationTotal', () => {
  it('returns 0 for empty input', () => {
    expect(computeDenominationTotal({})).toBe(0);
  });

  it('sums known denominations by count', () => {
    // 2 x 5000 rubles (500000 minor) + 3 x 100 rubles (10000 minor)
    expect(computeDenominationTotal({ 500000: '2', 10000: '3' })).toBe(1_030_000);
  });

  it('treats a zero count as contributing nothing', () => {
    expect(computeDenominationTotal({ 10000: '0' })).toBe(0);
  });

  it('treats a non-numeric count as contributing nothing', () => {
    expect(computeDenominationTotal({ 10000: 'abc' })).toBe(0);
  });

  it('characterization: a negative count is NOT ignored, it subtracts from the total', () => {
    // Numeric coercion via `Number(countStr) || 0` only falls back to 0 for
    // NaN/0 inputs; a negative numeric string is truthy and passes through,
    // so a negative count still contributes (denom * count) to the sum.
    expect(computeDenominationTotal({ 10000: '-1' })).toBe(-10_000);
  });

  it('adds the otherAmountMinor on top of the denomination sum', () => {
    expect(computeDenominationTotal({ 10000: '2' }, 500)).toBe(20_500);
  });
});

describe('actorsMustDiffer', () => {
  it('returns false when the actor and approver ids are the same', () => {
    expect(actorsMustDiffer('staff-1', 'staff-1')).toBe(false);
  });

  it('returns true when the actor and approver ids differ', () => {
    expect(actorsMustDiffer('staff-1', 'staff-2')).toBe(true);
  });

  it('treats two empty ids as equal (false)', () => {
    expect(actorsMustDiffer('', '')).toBe(false);
  });

  it('treats an empty id against a non-empty id as differing (true)', () => {
    expect(actorsMustDiffer('', 'staff-2')).toBe(true);
  });
});

describe('selectSuccessData', () => {
  it('returns the data payload for a 200 response', () => {
    const payload = { total: 42 };
    expect(selectSuccessData<typeof payload>({ status: 200, data: payload })).toEqual(payload);
  });

  it('returns undefined for a non-200 response', () => {
    expect(selectSuccessData({ status: 409, data: { code: 'conflict' } })).toBeUndefined();
  });

  it('returns undefined for an undefined response', () => {
    expect(selectSuccessData(undefined)).toBeUndefined();
  });

  it('returns undefined for a null response', () => {
    expect(selectSuccessData(null)).toBeUndefined();
  });
});
