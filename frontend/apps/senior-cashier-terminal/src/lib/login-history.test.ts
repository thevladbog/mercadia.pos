import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { deriveRecentLoginRecency, getRecentLogins, recordLogin } from './login-history.js';

const KEY = 'mercadia.sr-terminal.recent-logins';

beforeEach(() => {
  localStorage.clear();
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('recordLogin / getRecentLogins', () => {
  it('records a login retrievable via getRecentLogins', () => {
    recordLogin('senior-1', '2026-07-05T10:00:00.000Z');
    expect(getRecentLogins()).toEqual([{ actorId: 'senior-1', atIso: '2026-07-05T10:00:00.000Z' }]);
  });

  it('orders entries newest first', () => {
    recordLogin('senior-1', '2026-07-05T10:00:00.000Z');
    recordLogin('senior-2', '2026-07-05T11:00:00.000Z');
    expect(getRecentLogins().map((entry) => entry.actorId)).toEqual(['senior-2', 'senior-1']);
  });

  it('caps the stored history at 4 entries, dropping the oldest', () => {
    recordLogin('a', '2026-07-05T10:00:00.000Z');
    recordLogin('b', '2026-07-05T11:00:00.000Z');
    recordLogin('c', '2026-07-05T12:00:00.000Z');
    recordLogin('d', '2026-07-05T13:00:00.000Z');
    recordLogin('e', '2026-07-05T14:00:00.000Z');
    const ids = getRecentLogins().map((entry) => entry.actorId);
    expect(ids).toEqual(['e', 'd', 'c', 'b']);
    expect(ids).toHaveLength(4);
  });

  it('returns [] when nothing has been recorded', () => {
    expect(getRecentLogins()).toEqual([]);
  });

  it('falls back to [] when localStorage.getItem throws (unavailable storage)', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('storage unavailable');
    });
    expect(getRecentLogins()).toEqual([]);
  });

  it('silently no-ops when localStorage.setItem throws, without raising to the caller', () => {
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('quota exceeded');
    });
    expect(() => recordLogin('senior-1')).not.toThrow();
  });

  it('falls back to [] for corrupted JSON in the stored key', () => {
    localStorage.setItem(KEY, '{not valid json');
    expect(getRecentLogins()).toEqual([]);
  });

  it('filters out malformed entries (wrong shape) instead of throwing', () => {
    localStorage.setItem(
      KEY,
      JSON.stringify([
        { actorId: 'ok', atIso: '2026-07-05T10:00:00.000Z' },
        { bogus: true },
        'nope',
      ]),
    );
    expect(getRecentLogins()).toEqual([{ actorId: 'ok', atIso: '2026-07-05T10:00:00.000Z' }]);
  });
});

describe('deriveRecentLoginRecency', () => {
  const now = new Date(2026, 6, 5, 14, 30, 0).getTime(); // 2026-07-05 14:30 local

  it('buckets a login within the last 60s as "now"', () => {
    const atIso = new Date(now - 10_000).toISOString();
    expect(deriveRecentLoginRecency(atIso, now)).toEqual({ kind: 'now' });
  });

  it('buckets a same-day login older than 60s as a formatted HH:MM time', () => {
    const at = new Date(2026, 6, 5, 13, 18, 0);
    expect(deriveRecentLoginRecency(at.toISOString(), now)).toEqual({
      kind: 'time',
      hhmm: '13:18',
    });
  });

  it('buckets a login from an earlier calendar day as "earlier"', () => {
    const at = new Date(2026, 6, 4, 9, 0, 0);
    expect(deriveRecentLoginRecency(at.toISOString(), now)).toEqual({ kind: 'earlier' });
  });

  it('buckets an unparseable timestamp as "earlier" rather than throwing', () => {
    expect(deriveRecentLoginRecency('not-a-date', now)).toEqual({ kind: 'earlier' });
  });
});
