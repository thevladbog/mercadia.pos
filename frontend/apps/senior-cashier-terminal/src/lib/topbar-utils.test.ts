import { describe, expect, it } from 'vitest';

import {
  SESSION_TTL_MS,
  deriveElapsed,
  deriveInitials,
  deriveLoginAt,
  formatCountdown,
  formatLoginTime,
} from './topbar-utils.js';

describe('deriveInitials', () => {
  it('takes the first 2 alphanumeric characters, uppercased', () => {
    expect(deriveInitials('senior-1')).toBe('SE');
  });

  it('skips a leading non-alphanumeric separator', () => {
    expect(deriveInitials('-ab')).toBe('AB');
  });

  it('returns a single character when the id is only 1 character', () => {
    expect(deriveInitials('a')).toBe('A');
  });

  it('falls back to "?" for an id with no alphanumeric characters', () => {
    expect(deriveInitials('--')).toBe('?');
  });

  it('falls back to "?" for an empty id', () => {
    expect(deriveInitials('')).toBe('?');
  });
});

describe('deriveLoginAt', () => {
  it('subtracts the default 12h TTL from expiresAt', () => {
    const expiresAt = new Date(2026, 0, 1, 20, 0, 0).getTime();
    const loginAt = deriveLoginAt(new Date(expiresAt).toISOString());
    expect(expiresAt - loginAt).toBe(SESSION_TTL_MS);
  });

  it('supports a custom ttl override', () => {
    const expiresAt = new Date(2026, 0, 1, 9, 0, 0).getTime();
    const loginAt = deriveLoginAt(new Date(expiresAt).toISOString(), 60 * 60 * 1000);
    expect(expiresAt - loginAt).toBe(60 * 60 * 1000);
  });
});

describe('formatLoginTime', () => {
  it('formats a timestamp as 24-hour HH:MM', () => {
    const t = new Date(2026, 0, 1, 8, 5, 0).getTime();
    expect(formatLoginTime(t)).toBe('08:05');
  });

  it('zero-pads single-digit hours and minutes', () => {
    const t = new Date(2026, 0, 1, 0, 9, 0).getTime();
    expect(formatLoginTime(t)).toBe('00:09');
  });
});

describe('deriveElapsed', () => {
  it('splits elapsed time into whole hours and remainder minutes', () => {
    const loginAt = 0;
    const now = (3 * 60 + 12) * 60_000; // 3h 12m later
    expect(deriveElapsed(loginAt, now)).toEqual({ hours: 3, minutes: 12 });
  });

  it('clamps to zero when now is before loginAt', () => {
    expect(deriveElapsed(10_000, 0)).toEqual({ hours: 0, minutes: 0 });
  });
});

describe('formatCountdown', () => {
  it('formats sub-minute remaining time as 0:SS', () => {
    expect(formatCountdown(47_000)).toBe('0:47');
  });

  it('formats remaining time under an hour as M:SS', () => {
    expect(formatCountdown(5 * 60_000 + 3_000)).toBe('5:03');
  });

  it('formats remaining time at or over an hour as H:MM (seconds dropped)', () => {
    expect(formatCountdown(3 * 60 * 60_000 + 15 * 60_000 + 59_000)).toBe('3:15');
  });

  it('clamps negative remaining time to 0:00', () => {
    expect(formatCountdown(-5000)).toBe('0:00');
  });
});
