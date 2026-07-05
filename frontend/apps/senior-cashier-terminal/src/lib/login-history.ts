/**
 * Per-terminal local "recently logged in" history (plan 020).
 *
 * SCOPING — read before extending this: there is no cross-terminal or
 * store-wide login-history endpoint in the API. This is derived entirely
 * client-side, from successful logins performed ON THIS PHYSICAL TERMINAL
 * only. It is NOT a "who's currently on shift store-wide" feature. If a
 * future phase wants that, it needs a real backend endpoint (new query),
 * not an extension of this helper — see plan 020's maintenance notes.
 *
 * Deliberately uses `localStorage`, not `sessionStorage`: the active
 * session (`mercadia.sr-terminal.session` in `AuthProvider.tsx`) and the
 * failed-attempts counter (`mercadia.sr-terminal.login-attempts` in
 * `LoginPage.tsx`) already live in `sessionStorage` under their own keys,
 * and both are meant to reset per-session. This history must survive
 * logout and across sessions on this device, hence `localStorage`.
 */

const RECENT_LOGINS_KEY = 'mercadia.sr-terminal.recent-logins';
const MAX_RECENT_LOGINS = 4;

export type RecentLogin = {
  actorId: string;
  atIso: string;
};

function isRecentLogin(value: unknown): value is RecentLogin {
  if (typeof value !== 'object' || value === null) return false;
  const candidate = value as Partial<RecentLogin>;
  return typeof candidate.actorId === 'string' && typeof candidate.atIso === 'string';
}

/**
 * Read the local login history, newest first. Defensively falls back to
 * `[]` on any failure (localStorage unavailable, corrupted JSON, wrong
 * shape, etc.) — same defensive try/catch-to-empty style as the existing
 * `loadAttempts` in the pre-wizard `LoginPage.tsx`.
 */
export function getRecentLogins(): RecentLogin[] {
  try {
    const raw = localStorage.getItem(RECENT_LOGINS_KEY);
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter(isRecentLogin);
  } catch {
    return [];
  }
}

/**
 * Record a successful login: prepend a new entry (newest first) and cap the
 * stored list at the last `MAX_RECENT_LOGINS` entries. `atIso` is
 * injectable for testability; defaults to "now". Silently no-ops on any
 * localStorage failure — this is a UX nicety, not part of the auth
 * critical path, so it must never throw into the caller.
 */
export function recordLogin(actorId: string, atIso: string = new Date().toISOString()): void {
  try {
    const next = [{ actorId, atIso }, ...getRecentLogins()].slice(0, MAX_RECENT_LOGINS);
    localStorage.setItem(RECENT_LOGINS_KEY, JSON.stringify(next));
  } catch {
    /* noop */
  }
}

/**
 * Recency bucket for a recent-login chip's caption, matching design screen
 * 01a's examples ("сейчас" / "13:18" / "вчера"). Returns a discriminated
 * union rather than a formatted string so the actual wording stays in
 * i18n (`auth.wizard.*`), matching how `topbar-utils.ts`'s
 * `formatRoleLabel` takes a `translate` function instead of hardcoding
 * text.
 *
 * Judgment call: the design only shows two concrete buckets (same-day
 * "now"/HH:MM and a generic "earlier" for anything before today); this
 * collapses everything before the current calendar day into one
 * `'earlier'` bucket rather than computing "yesterday" vs. "2 days ago"
 * distinctions, since the list is capped at 4 entries and multi-day-old
 * entries are the rare case here.
 */
export type RecentLoginRecency =
  { kind: 'now' } | { kind: 'time'; hhmm: string } | { kind: 'earlier' };

export function deriveRecentLoginRecency(
  atIso: string,
  nowMs: number = Date.now(),
): RecentLoginRecency {
  const atMs = new Date(atIso).getTime();
  if (!Number.isFinite(atMs)) return { kind: 'earlier' };

  const diffMs = nowMs - atMs;
  if (diffMs < 60_000 && diffMs >= 0) return { kind: 'now' };

  const at = new Date(atMs);
  const now = new Date(nowMs);
  const sameDay =
    at.getFullYear() === now.getFullYear() &&
    at.getMonth() === now.getMonth() &&
    at.getDate() === now.getDate();
  if (sameDay) {
    const hh = String(at.getHours()).padStart(2, '0');
    const mm = String(at.getMinutes()).padStart(2, '0');
    return { kind: 'time', hhmm: `${hh}:${mm}` };
  }
  return { kind: 'earlier' };
}
