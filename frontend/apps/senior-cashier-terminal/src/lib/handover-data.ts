/**
 * Pure join/derivation helpers backing the redesigned `ShiftHandoverPage`
 * (plan 029, Phase 9). Mirrors `dashboard-data.ts`'s exact conventions
 * (narrow local interfaces per source endpoint, pure functions, no DOM/hooks,
 * colocated and unit-tested the same way) — see that file's own top comment
 * for the shared "no DOM in `lib/`" discipline this module follows too.
 *
 * SCOPING — read before extending this file: plan 029's "Ground truth"
 * section documents, per cash-operation page, exactly which
 * `CashMovementType` maps to which session-summary category and which
 * identity field is reliable to filter on. Three of the five
 * cash-operation pages (`IssueChangeFundPage`, `ReceiveCashPage`,
 * `FinalCollectionPage`) auto-derive `actorId`/`approvedById` from a
 * selected shift + this terminal's own session, so `approvedById` alone
 * would be a reliable match there. The other two (`BankCollectionPage`,
 * `BusinessExpensePage`) keep `actorId`/`approvedById` as manual free-text
 * inputs (plan 022), so `approvedById` alone is less reliable — an operator
 * could type any value into either field. Per plan 029, this module filters
 * uniformly by `movement.actorId === actorId || movement.approvedById ===
 * actorId` for every movement-based category, since that's a strict
 * superset of "approvedById only" (the 3 auto-derived pages additionally
 * enforce `actorsMustDiffer(actorId, approvedById)`, so their own `actorId`
 * — the shift's cashier — can never coincidentally equal this terminal's own
 * signed-in senior cashier) while correctly covering the 2 manual-entry
 * pages too. The small risk of over-inclusion this uniform filter carries
 * applies only to those 2 manual-entry types — a known limitation inherited
 * from their own design (plan 022), not something this module can fix.
 *
 * "Refunds approved" is deliberately NOT a `CashMovement` category at all —
 * it comes from `Return.ConfirmPendingApproval` (plan 025) via
 * `listStoreReturns`, filtered by a real, non-zero `confirmedAt` (not
 * `status === 'completed'`): a with-receipt return is `completed`
 * immediately at creation with `confirmedAt` staying the Go zero value
 * (serializes as `"0001-01-01T00:00:00Z"`, well before the Unix epoch), and
 * would otherwise be miscounted as something this senior cashier "approved"
 * when it never went through the confirm flow at all.
 */

import type { CredentialActorForJoin } from './dashboard-data.js';

export type { CredentialActorForJoin };

/**
 * One row of the successor picker (plan 029, `ShiftHandoverPage`'s right
 * panel). `role` is `string | null` for the same reason
 * `dashboard-data.ts`'s `CashierOnShift.role` is — no actor is guaranteed to
 * have any roles at all, and this reuses `formatRoleLabel`'s actual
 * parameter type instead of inventing a parallel `Role` union.
 */
export interface EligibleSuccessor {
  actorId: string;
  role: string | null;
}

/**
 * Real membership for the successor picker: every credential-management
 * actor with `senior_cashier` or `admin` among their roles, excluding the
 * CLOSING senior cashier's own `actorId` (can't hand off to yourself). Per
 * plan 029's ground truth, the design mockup's status badges
 * (`available`/`on-shift`/`remote`) and differentiated role labels have zero
 * backing data anywhere in this backend — this only returns the real
 * `actorId` + real first role, never a fabricated status.
 */
export function joinEligibleSuccessors(
  actors: CredentialActorForJoin[],
  excludeActorId: string,
): EligibleSuccessor[] {
  return actors
    .filter((actor) => actor.id !== excludeActorId)
    .filter((actor) => actor.roles.includes('senior_cashier') || actor.roles.includes('admin'))
    .map((actor) => ({ actorId: actor.id, role: actor.roles[0] ?? null }));
}

/**
 * Minimal shape needed from `ListCashMovements200ItemsItem`
 * (`frontend/packages/api-clients/store-edge/src/generated/models/listCashMovements200ItemsItem.ts:10-26`).
 */
export interface CashMovementForJoin {
  type: string;
  fromContainerType: string;
  toContainerType: string;
  amountMinor: number;
  actorId: string;
  approvedById?: string;
  createdAt: string;
}

/**
 * Minimal shape needed from `ListStoreReturns200ItemsItem`
 * (`frontend/packages/api-clients/store-edge/src/generated/models/listStoreReturns200ItemsItem.ts:10-23`).
 * `totalMinor` is included alongside `approvedById`/`confirmedAt` — without
 * it there would be no real figure to sum into `refundsApprovedMinor` below.
 */
export interface ReturnForJoin {
  approvedById?: string;
  confirmedAt?: string;
  totalMinor: number;
}

/**
 * The 6-row session summary (5 real financial categories + the safe net
 * change aggregate). Each financial row carries its own op count alongside
 * its amount, for the "N операций" sub-label the design shows under every
 * row — not just the one financial category (refunds) whose amount can't be
 * derived from a plain `.length` the way the movement-based ones can.
 */
export interface HandoverSummary {
  changeIssuedMinor: number;
  changeIssuedCount: number;
  cashReceivedMinor: number;
  cashReceivedCount: number;
  incassationsMinor: number;
  incassationsCount: number;
  expensesMinor: number;
  expensesCount: number;
  refundsApprovedMinor: number;
  refundsApprovedCount: number;
  safeNetChangeMinor: number;
  safeNetChangeCount: number;
}

/** See this file's top doc comment on why this OR-condition is applied
 * uniformly across every movement-based category. */
function matchesActor(movement: CashMovementForJoin, actorId: string): boolean {
  return movement.actorId === actorId || movement.approvedById === actorId;
}

function sumMovementsByType(
  movements: CashMovementForJoin[],
  type: string,
  actorId: string,
  sinceIso: string,
): { totalMinor: number; count: number } {
  const matches = movements.filter(
    (movement) =>
      movement.type === type && matchesActor(movement, actorId) && movement.createdAt >= sinceIso,
  );
  return {
    totalMinor: matches.reduce((sum, movement) => sum + movement.amountMinor, 0),
    count: matches.length,
  };
}

/**
 * A return counts as "approved" for this session only once it has gone
 * through the real `ConfirmPendingApproval` flow — see this file's top doc
 * comment on why `confirmedAt` (not `status`) is the right signal, and why a
 * Go zero-value timestamp (parses to a large negative number, well before
 * the 1970 epoch) must be treated as "not confirmed."
 */
function isApprovedReturnSince(
  returnItem: ReturnForJoin,
  actorId: string,
  sinceIso: string,
): boolean {
  if (returnItem.approvedById !== actorId) return false;
  if (!returnItem.confirmedAt) return false;
  const confirmedAtMs = new Date(returnItem.confirmedAt).getTime();
  if (!Number.isFinite(confirmedAtMs) || confirmedAtMs <= 0) return false;
  return returnItem.confirmedAt >= sinceIso;
}

/**
 * Computes the closing senior cashier's session-summary figures, per plan
 * 029's ground-truth field-mapping table:
 * - "Выдано размена" (change issued): `type === 'change_fund'`
 *   (`IssueChangeFundPage.tsx`).
 * - "Принято от кассиров" (cash received): `type === 'cash_out'`
 *   (`ReceiveCashPage.tsx:123-124` — named from the cashier's drawer
 *   perspective, confirmed in `domain/cash.go:20`).
 * - "Инкассации (итог)" (final incassations): `type === 'drawer_to_safe'`
 *   (`FinalCollectionPage.tsx`'s `closeShift`-internal collection movement,
 *   atomic fast path or the plan-026 two-stage `confirmCloseShift` flow).
 * - "Хоз. расход" (expenses): `type === 'expense'` (`BusinessExpensePage.tsx`
 *   via `createBusinessExpense`). Note `BankCollectionPage.tsx`'s
 *   `safe_to_bank` movements have no row of their own in this 6-row design —
 *   they only ever contribute to `safeNetChangeMinor` below.
 * - "Возвраты подтверждено" (refunds approved): NOT a `CashMovement` at all
 *   — see this file's top doc comment.
 * - "Сейф · итог изменения" (safe net change): signed sum across every
 *   safe-touching movement matching this session's actor filter
 *   (`toContainerType === 'safe'` adds, `fromContainerType === 'safe'`
 *   subtracts) — a genuinely computable aggregate, not a separate category,
 *   so it can include `safe_to_bank`/any other type that touches the safe.
 *
 * Every category is additionally bounded to `createdAt`/`confirmedAt >=
 * sinceIso` (the session's `loggedInAt`) — this is a SESSION summary, not an
 * all-time one.
 */
export function deriveHandoverSummary(
  movements: CashMovementForJoin[],
  returns: ReturnForJoin[],
  actorId: string,
  sinceIso: string,
): HandoverSummary {
  const changeIssued = sumMovementsByType(movements, 'change_fund', actorId, sinceIso);
  const cashReceived = sumMovementsByType(movements, 'cash_out', actorId, sinceIso);
  const incassations = sumMovementsByType(movements, 'drawer_to_safe', actorId, sinceIso);
  const expenses = sumMovementsByType(movements, 'expense', actorId, sinceIso);

  const approvedReturns = returns.filter((returnItem) =>
    isApprovedReturnSince(returnItem, actorId, sinceIso),
  );
  const refundsApprovedMinor = approvedReturns.reduce(
    (sum, returnItem) => sum + returnItem.totalMinor,
    0,
  );

  const safeTouchingMovements = movements.filter(
    (movement) =>
      matchesActor(movement, actorId) &&
      movement.createdAt >= sinceIso &&
      (movement.fromContainerType === 'safe' || movement.toContainerType === 'safe'),
  );
  const safeNetChangeMinor = safeTouchingMovements.reduce((sum, movement) => {
    if (movement.toContainerType === 'safe') return sum + movement.amountMinor;
    if (movement.fromContainerType === 'safe') return sum - movement.amountMinor;
    return sum;
  }, 0);

  return {
    changeIssuedMinor: changeIssued.totalMinor,
    changeIssuedCount: changeIssued.count,
    cashReceivedMinor: cashReceived.totalMinor,
    cashReceivedCount: cashReceived.count,
    incassationsMinor: incassations.totalMinor,
    incassationsCount: incassations.count,
    expensesMinor: expenses.totalMinor,
    expensesCount: expenses.count,
    refundsApprovedMinor,
    refundsApprovedCount: approvedReturns.length,
    safeNetChangeMinor,
    safeNetChangeCount: safeTouchingMovements.length,
  };
}

/** Carry-over warning counts (plan 029). Both are store-wide, real pending
 * states — `domain.CashMovementStatus` has exactly one value (`posted`), so
 * there is no "open incassation" to count the way the design mockup implies;
 * these two ARE genuinely open/pending today. */
export interface HandoverWarnings {
  pendingChangeFundRequests: number;
  openRecountDiscrepancies: number;
}

/** Minimal shape needed from `ListStoreChangeFundRequests200ItemsItem`'s
 * `status` field (`ChangeFundRequestStatusRequested`,
 * `domain/change_fund_request.go:11`). */
export interface ChangeFundRequestForJoin {
  status: string;
}

/** Minimal shape needed from `ListCashRecounts200ItemsItem`'s
 * `resolutionStatus` field (`CashRecountResolutionStatusOpen`,
 * `domain/cash_recount.go:16`). */
export interface CashRecountForJoin {
  resolutionStatus: string;
}

/**
 * Counts of the 2 real pending states available store-wide (plan 029, user
 * decision #3): unfulfilled change-fund requests (plan 027) and unresolved
 * cash-recount discrepancies (plan 028). Neither resource is scoped to this
 * senior-cashier terminal specifically, so both counts are honestly
 * store-wide, not lane-specific.
 */
export function deriveHandoverWarnings(
  changeFundRequests: ChangeFundRequestForJoin[],
  cashRecounts: CashRecountForJoin[],
): HandoverWarnings {
  return {
    pendingChangeFundRequests: changeFundRequests.filter(
      (request) => request.status === 'requested',
    ).length,
    openRecountDiscrepancies: cashRecounts.filter((recount) => recount.resolutionStatus === 'open')
      .length,
  };
}
