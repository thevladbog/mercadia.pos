/**
 * Pure derivation helpers backing the redesigned cash-operation pages (plan
 * 022, Phase 5): `IssueChangeFundPage`, `ReceiveCashPage`,
 * `FinalCollectionPage`, `BankCollectionPage`, `BusinessExpensePage`. Kept
 * side-effect free and colocated under `pages/cash-operations/`, mirroring
 * `lib/dashboard-data.ts`'s "no DOM in the join/derivation layer" convention
 * (plan 021).
 *
 * SCOPING — read before extending this file: plan 022's "Why this matters"
 * documents two real pre-existing bugs this module exists to fix (items 3
 * and 4). `FinalCollectionPage`'s mismatch check compared the operator's
 * count against a hardcoded `0` (`selectedShift?.closingCashMinor ?? 0`,
 * which is the Go zero-value for every OPEN shift) instead of the shift's
 * real current drawer balance. And every "before" balance lookup across
 * these pages picked the FIRST cash balance of a given `containerType`
 * rather than the one belonging to the actually selected shift's own
 * `drawerId` — silently showing/using the wrong container whenever more
 * than one shift (more than one drawer) is open. Nothing here invents data:
 * every helper only recombines fields that genuinely exist on
 * `useListCashBalances`/`useListOpenStoreShifts`/`useGetCredentialManagement`.
 */

/**
 * Minimal shape needed from `ListCashBalances200BalancesItem`
 * (`frontend/packages/api-clients/store-edge/src/generated/models/listCashBalances200BalancesItem.ts:9-16`).
 */
export interface CashBalanceForLookup {
  containerId: string;
  containerType: string;
  balanceMinor: number;
}

/**
 * Find a container's cash balance by its exact `containerId` — never falls
 * back to "first balance of this containerType". Plan 022 item 4: with more
 * than one open shift there is more than one `drawer` container, and a bare
 * `.find(b => b.containerType === 'drawer')` silently returns an arbitrary
 * OTHER shift's drawer. Every shift-scoped "before" balance lookup in these
 * pages must go through this helper with the selected shift's own
 * `drawerId`, never a bare containerType match.
 */
export function findContainerBalance(
  balances: CashBalanceForLookup[],
  containerId: string | undefined,
): CashBalanceForLookup | undefined {
  if (!containerId) return undefined;
  return balances.find((balance) => balance.containerId === containerId);
}

/**
 * Find the store's single safe-container balance. Unlike drawers (one per
 * open shift, so several can coexist — see `findContainerBalance` above),
 * the backend models exactly one safe per store today; every cash-operation
 * page already assumes this (e.g. the hardcoded `safe-1` fallback ids on
 * Bank Collection / Business Expense), so a plain `containerType` match is
 * not the same "wrong container" bug pattern as the drawer case.
 */
export function findSafeBalance(
  balances: CashBalanceForLookup[],
): CashBalanceForLookup | undefined {
  return balances.find((balance) => balance.containerType === 'safe');
}

export type BalanceDirection = 'increase' | 'decrease';

/**
 * Client-side "after" preview for `BeforeAfterPanel` — `before ± total`.
 * This is NOT a server-confirmed post-transaction figure: a concurrent
 * operation could change the real balance before this one is submitted.
 * Mirrors the same "optimistic UI, not a guarantee" framing plan 020
 * already established for its step-checkmarks.
 */
export function computeAfterBalance(
  beforeMinor: number,
  totalMinor: number,
  direction: BalanceDirection,
): number {
  return direction === 'increase' ? beforeMinor + totalMinor : beforeMinor - totalMinor;
}

/**
 * Minimal shape needed from `GetCredentialManagement200ActorsItem`. Kept as
 * its own local interface (rather than importing `dashboard-data.ts`'s
 * `CredentialActorForJoin`) so this module has no cross-feature import
 * dependency, matching how `dashboard-data.ts` itself declares narrow inline
 * shapes instead of importing generated models.
 */
export interface CredentialActorForRoleLookup {
  id: string;
  roles: string[];
}

/**
 * Resolve a shift's cashier's role for `CashierIdentityBanner` — the SAME
 * join precedent `dashboard-data.ts`'s `joinCashiersOnShift` already
 * established (`actor.roles[0]`, `null` if no actor matches or the actor
 * has no roles), not a second role-lookup mechanism.
 */
export function findActorRole(
  actors: CredentialActorForRoleLookup[],
  actorId: string | undefined,
): string | null {
  if (!actorId) return null;
  const actor = actors.find((item) => item.id === actorId);
  return actor?.roles?.[0] ?? null;
}
