/**
 * Pure derivation helpers backing the `TopBar` component (plans/019). Kept
 * side-effect free and colocated with `cash-utils.ts` so the identity/
 * countdown math is unit-testable without rendering, mirroring how
 * `apps/sco-terminal/src/stage.ts` keeps its state machine pure alongside
 * the component that consumes it.
 */

/**
 * The backend hardcodes a 12-hour session TTL at creation time (confirmed at
 * `backend/services/store-edge/internal/app/auth.go:159` and
 * `backend/services/store-edge/internal/domain/session.go:89`). There is no
 * login timestamp on the session payload, so it must be derived from
 * `expiresAt - ttl`.
 */
export const SESSION_TTL_MS = 12 * 60 * 60 * 1000;

/**
 * Derive 1-2 letter initials from an actor id for `AvatarChip`.
 * Rule: take the first two alphanumeric characters of the id, uppercased
 * (e.g. "senior-1" -> "SE", "a" -> "A", "" -> "?"). Simple and deterministic;
 * no name field exists on `domain.Actor` to parse instead (see plan 019
 * "Current state").
 */
export function deriveInitials(actorId: string): string {
  const alnum = actorId.match(/[a-zA-Z0-9]/g) ?? [];
  if (alnum.length === 0) return '?';
  return alnum.slice(0, 2).join('').toUpperCase();
}

/** Derive the session's login timestamp (epoch ms) from `expiresAt` and the known TTL. */
export function deriveLoginAt(expiresAtIso: string, ttlMs: number = SESSION_TTL_MS): number {
  return new Date(expiresAtIso).getTime() - ttlMs;
}

/** Format a login timestamp as a 24-hour local `HH:MM` string. */
export function formatLoginTime(loginAtMs: number): string {
  const d = new Date(loginAtMs);
  const hh = String(d.getHours()).padStart(2, '0');
  const mm = String(d.getMinutes()).padStart(2, '0');
  return `${hh}:${mm}`;
}

/** Elapsed session duration, broken into whole hours + remainder minutes. */
export function deriveElapsed(
  loginAtMs: number,
  nowMs: number,
): { hours: number; minutes: number } {
  const totalMinutes = Math.max(0, Math.floor((nowMs - loginAtMs) / 60_000));
  return { hours: Math.floor(totalMinutes / 60), minutes: totalMinutes % 60 };
}

/**
 * Format the auto-lock countdown to match the design's `0:47` style: once
 * under an hour remains, show `M:SS`; at an hour or more, show `H:MM`
 * (seconds dropped — the design only shows a sub-hour example, so the >1h
 * format is a judgment call favoring the same two-segment shape).
 */
export function formatCountdown(remainingMs: number): string {
  const totalSeconds = Math.max(0, Math.floor(remainingMs / 1000));
  const hours = Math.floor(totalSeconds / 3600);
  if (hours >= 1) {
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    return `${hours}:${String(minutes).padStart(2, '0')}`;
  }
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes}:${String(seconds).padStart(2, '0')}`;
}

/**
 * Maps a raw role enum value (`cashier` | `senior_cashier` | `admin`, per
 * `backend/services/store-edge/internal/domain/session.go:11-13`) to its
 * i18n key for a human-readable label. The actual label strings live in
 * `i18n/locales/*.json` under `topbar.roles.*`, not here — `translate` is
 * injected so this stays a pure, testable function without needing an i18n
 * context in tests, matching how the rest of TopBar sources its strings.
 */
const ROLE_LABEL_KEYS: Record<string, string> = {
  cashier: 'topbar.roles.cashier',
  senior_cashier: 'topbar.roles.seniorCashier',
  admin: 'topbar.roles.admin',
};

/** Human-readable label for a role; falls back to the raw value for any unrecognized role. */
export function formatRoleLabel(role: string, translate: (key: string) => string): string {
  const key = ROLE_LABEL_KEYS[role];
  return key ? translate(key) : role;
}
