/**
 * Pure join/derivation helpers backing the redesigned `DashboardPage`
 * (plan 021, Phase 4). Kept side-effect free and colocated with
 * `topbar-utils.ts`/`login-wizard.ts`/`login-history.ts` so every
 * cross-endpoint join and count derivation is unit-testable without
 * rendering — same "no DOM in `lib/`" convention as those modules.
 *
 * SCOPING — read before extending this file: plan 021's "Why this matters"
 * section documents 9 real backend gaps (no cashier display name, no safe
 * limit, no return-pending state, etc.). Nothing in this file invents data
 * to paper over those gaps — every join here only recombines fields that
 * genuinely exist on the three source endpoints (open shifts, monitoring
 * terminals, credential management), and every derived count/label is
 * computed from real fields, not fabricated. See the plan for the full
 * rationale before adding anything that looks like a shortcut around a gap.
 */

/**
 * Minimal shape needed from `ListOpenStoreShifts200ShiftsItem`
 * (`frontend/packages/api-clients/store-edge/src/generated/models/listOpenStoreShifts200ShiftsItem.ts:9-23`).
 * A narrow local interface rather than the full generated type, mirroring
 * how `DashboardPage.tsx`/`OperationJournalPage.tsx` already declare their
 * own minimal inline response shapes instead of importing generated models
 * into this dependency-free `lib/` layer.
 */
export interface OpenShiftForJoin {
  id: string;
  cashierId: string;
  terminalId: string;
  openedAt: string;
}

/**
 * Minimal shape needed from `ListStoreMonitoringTerminals200ItemsItem`
 * (`frontend/packages/api-clients/store-edge/src/generated/models/listStoreMonitoringTerminals200ItemsItem.ts:9-27`).
 * This is the ONLY source of real per-terminal `revenueMinor`/
 * `drawerBalanceMinor` — the open-shifts endpoint's `closingCashMinor` is a
 * close-time field that stays `0` for open shifts and must not be used for
 * "выручка" (see plan 021 "Why this matters" item 3).
 */
export interface MonitoringTerminalForJoin {
  id: string;
  revenueMinor: number;
  drawerBalanceMinor: number;
}

/**
 * Minimal shape needed from `GetCredentialManagement200ActorsItem`
 * (`frontend/packages/api-clients/store-edge/src/generated/models/getCredentialManagement200ActorsItem.ts:9-13`).
 */
export interface CredentialActorForJoin {
  id: string;
  roles: string[];
}

/**
 * Minimal shape needed from `ListOperationJournal200ItemsItem`
 * (`frontend/packages/api-clients/store-edge/src/generated/models/listOperationJournal200ItemsItem.ts:9-17`).
 */
export interface OperationJournalEntryForJoin {
  actorId: string;
  createdAt: string;
}

/**
 * One row of the "Кассиры на смене" panel. `role` is `string | null` rather
 * than a closed union — this reuses `formatRoleLabel`'s actual parameter
 * type (`role: string`) from `topbar-utils.ts` verbatim instead of inventing
 * a parallel `Role` union, per plan 021 step 2's explicit instruction.
 */
export interface CashierOnShift {
  shiftId: string;
  cashierId: string;
  role: string | null;
  terminalId: string;
  openedAt: string;
  revenueMinor: number;
  drawerBalanceMinor: number;
}

/**
 * Join open shifts with their live monitoring-terminal figures and the
 * cashier's role, for the "Кассиры на смене" panel.
 *
 * - Terminal match: `monitoringTerminal.id === shift.terminalId`. A terminal
 *   can be momentarily absent from the monitoring list (e.g. just opened,
 *   not yet polled) — this defaults `revenueMinor`/`drawerBalanceMinor` to
 *   `0` rather than throwing or dropping the row.
 * - Role match: `credentialActor.id === shift.cashierId`, using
 *   `actor.roles[0]`. If no actor matches, or the actor has no roles, `role`
 *   is `null` — callers must omit the role badge entirely rather than guess
 *   (plan 021 "Why this matters" item 1).
 *
 * Pure and hook-free so it can be unit tested directly against fixture
 * arrays; `DashboardPage.tsx` is responsible for calling the three
 * `useList*`/`useGet*` hooks and passing their `data` through.
 */
export function joinCashiersOnShift(
  shifts: OpenShiftForJoin[],
  monitoringTerminals: MonitoringTerminalForJoin[],
  credentialActors: CredentialActorForJoin[],
): CashierOnShift[] {
  return shifts.map((shift) => {
    const terminal = monitoringTerminals.find((item) => item.id === shift.terminalId);
    const actor = credentialActors.find((item) => item.id === shift.cashierId);
    const role = actor?.roles?.[0] ?? null;

    return {
      shiftId: shift.id,
      cashierId: shift.cashierId,
      role,
      terminalId: shift.terminalId,
      openedAt: shift.openedAt,
      revenueMinor: terminal?.revenueMinor ?? 0,
      drawerBalanceMinor: terminal?.drawerBalanceMinor ?? 0,
    };
  });
}

/**
 * Approximate "ОПЕРАЦИЙ N" count for the TopBar pill and the "Журнал смены"
 * card subtitle (plan 021 "Why this matters" item 8).
 *
 * There is no backend endpoint for "operations by this actor, on this
 * terminal, this session" — `useListOperationJournal` is store-wide and
 * unbounded by actor or terminal. This derives a real-data approximation
 * client-side: entries actually authored by `actorId`, with a real
 * `createdAt` at or after `sinceIso` (the session's derived login time).
 * It is NOT a true "operations this session on this terminal" count — the
 * journal spans every terminal this actor may have used since logging in,
 * and if a future phase adds a dedicated backend counter, that should
 * replace this derivation rather than stack more client-side filtering on
 * top of it (see plan 021's maintenance notes).
 */
export function deriveOperationsCount(
  journalEntries: OperationJournalEntryForJoin[],
  actorId: string,
  sinceIso: string,
): number {
  return journalEntries.filter((entry) => entry.actorId === actorId && entry.createdAt >= sinceIso)
    .length;
}

/** Minimal shape needed for `deriveAlertsCount`, matching the two fields of
 * `GetStoreMonitoringSummary200` it uses
 * (`frontend/packages/api-clients/store-edge/src/generated/models/getStoreMonitoringSummary200.ts:9-18`). */
export interface AlertsSummaryForDerive {
  attentionTerminalCount: number;
  offlineTerminalCount: number;
}

/**
 * "АЛЕРТЫ" count for the TopBar pill and the "Мониторинг касс" card, per
 * plan 021 "Why this matters" item 9 — this DOES have a direct real source,
 * unlike the operations count above. Terminals needing attention plus fully
 * offline terminals both count as "alerts" needing a senior cashier's eyes.
 */
export function deriveAlertsCount(summary: AlertsSummaryForDerive | undefined): number {
  return (summary?.attentionTerminalCount ?? 0) + (summary?.offlineTerminalCount ?? 0);
}

/** Minimal shape needed for `deriveTotalTerminalCount` — the 3 terminal-
 * count fields of `GetStoreMonitoringSummary200` it sums
 * (`frontend/packages/api-clients/store-edge/src/generated/models/getStoreMonitoringSummary200.ts:9-18`). */
export interface TerminalCountsSummaryForDerive {
  activeTerminalCount: number;
  freeTerminalCount: number;
  offlineTerminalCount: number;
}

/**
 * Total known terminals ("N узлов" in the "Мониторинг касс" card subtitle),
 * per plan 021 "Why this matters" item 9's exact formula: active + free +
 * offline. Returns `undefined` (not `0`) while the summary hasn't loaded
 * yet, so callers can distinguish "no data yet" from "genuinely zero
 * terminals" the same way the rest of this dashboard treats loading state.
 */
export function deriveTotalTerminalCount(
  summary: TerminalCountsSummaryForDerive | undefined,
): number | undefined {
  if (!summary) return undefined;
  return summary.activeTerminalCount + summary.freeTerminalCount + summary.offlineTerminalCount;
}

/**
 * Translate-function shape reused across this module's label helpers,
 * matching `formatRoleLabel`'s `(key) => string` shape but widened to
 * accept i18next-style interpolation params, since these labels need
 * `{{count}}`/`{{label}}` placeholders that `formatRoleLabel` doesn't.
 */
export type Translate = (key: string, params?: Record<string, unknown>) => string;

/**
 * Freshness label for the "Сейф · сейчас" panel's "обновлено N назад"
 * caption (design screen 02's "ОБНОВЛЕНО · 1 СЕК").
 *
 * Judgment call: `login-history.ts`'s `deriveRecentLoginRecency` buckets
 * into now/HH:MM/earlier because it labels a list of logins that can be
 * hours or days old. This panel instead re-renders continuously and the
 * design shows sub-minute freshness ("1 СЕК"), so a single bucketed
 * "earlier" state would look stale the moment more than a few minutes pass.
 * This returns a live seconds/minutes/hours-ago string instead — reads
 * better for a value that's expected to refresh constantly.
 *
 * Must not throw on a missing or unparseable timestamp (mirrors the
 * `Number.isFinite(atMs)` guard already used by `deriveRecentLoginRecency`).
 */
export function deriveSafeFreshnessLabel(
  lastMovementAtIso: string | undefined,
  nowMs: number,
  t: Translate,
): string {
  const atMs = lastMovementAtIso ? new Date(lastMovementAtIso).getTime() : NaN;
  if (!Number.isFinite(atMs)) return t('dashboard.safeNow.freshnessUnknown');

  const diffSec = Math.max(0, Math.floor((nowMs - atMs) / 1000));
  if (diffSec < 60) return t('dashboard.safeNow.freshnessSeconds', { count: diffSec });

  const diffMin = Math.floor(diffSec / 60);
  if (diffMin < 60) return t('dashboard.safeNow.freshnessMinutes', { count: diffMin });

  const diffHours = Math.floor(diffMin / 60);
  return t('dashboard.safeNow.freshnessHours', { count: diffHours });
}
