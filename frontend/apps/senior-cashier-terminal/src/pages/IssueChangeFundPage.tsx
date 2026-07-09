import { useMemo, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button } from '@mercadia/ui';
import {
  useListCashBalances,
  useListOpenStoreShifts,
  useGetCredentialManagement,
  createCashMovement,
  getListCashBalancesQueryKey,
} from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { useAuth } from '@/auth/AuthProvider.js';
import { getStoreId } from '@/api-client-config.js';
import {
  actorsMustDiffer,
  computeDenominationTotal,
  createIdempotencyHeaders,
  selectSuccessData,
} from '@/lib/cash-utils.js';
import { useTopBarActions } from '@/lib/use-topbar-actions.js';
import { CashierSelectModal } from '@/components/CashierSelectModal.js';
import { TopBar } from '@/components/TopBar.js';

import { DownArrowIcon } from './dashboard/icons.js';
import { BeforeAfterPanel } from './cash-operations/BeforeAfterPanel.js';
import { CashierIdentityBanner } from './cash-operations/CashierIdentityBanner.js';
import { DenominationGrid } from './cash-operations/DenominationGrid.js';
import { OperationChecklist } from './cash-operations/OperationChecklist.js';
import {
  findActorRole,
  findContainerBalance,
  findSafeBalance,
  type CashBalanceForLookup,
  type CredentialActorForRoleLookup,
} from './cash-operations/cash-operations-data.js';
import './cash-operations/CashOperations.css';

// Stable empty-array fallbacks (module scope, never reassigned) so the
// derived container/role lookups below don't see a fresh `[]` reference —
// and therefore a false "changed" input — on every render while a query
// hasn't resolved yet. Same convention as `dashboard-data.ts`'s
// `EMPTY_SHIFTS`/`EMPTY_TERMINALS`/`EMPTY_ACTORS` (plan 021).
const EMPTY_BALANCES: CashBalanceForLookup[] = [];
const EMPTY_ACTORS: CredentialActorForRoleLookup[] = [];

/**
 * Redesigned Issue Change Fund page (plan 022, Phase 5; design screen 03a).
 * A thin orchestrator delegating to the shared `pages/cash-operations/`
 * components, same discipline as plan 021's `DashboardPage.tsx`.
 *
 * `actorId`/`approvedById` are auto-derived (plan 022 item 5) — no manual
 * free-text inputs — since this page already has a `CashierSelectModal`:
 * the cashier from `selectedShift.cashierId`, the approver from this
 * terminal's own signed-in session (`useAuth()`'s `session.actorId`).
 *
 * No `MismatchDialog` — there is no "expected" figure to compare against
 * for issuing change fund (an operator-decided amount). The design's "Не
 * сошлось" note is kept as informational copy near the confirm button,
 * not wired to a modal.
 *
 * CodeRabbit fix (PR #86): submission is now blocked until the real safe
 * balance resolves — same "don't let `safeBalance?.containerId ?? 'safe-1'`
 * post against the placeholder id" guard `BankCollectionPage`/
 * `ReceiveCashPage` now have.
 */
export function IssueChangeFundPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { onHandover, onLock } = useTopBarActions();
  const { session } = useAuth();
  const storeId = useMemo(() => getStoreId(), []);

  const { data: balancesResp } = useListCashBalances(storeId);
  const { data: shiftsResp } = useListOpenStoreShifts(storeId);
  const { data: credentialsResp } = useGetCredentialManagement(storeId);

  const balances =
    selectSuccessData<{ balances: CashBalanceForLookup[] }>(balancesResp)?.balances ??
    EMPTY_BALANCES;
  const shiftsData = selectSuccessData<{
    shifts: { id: string; cashierId: string; drawerId: string }[];
  }>(shiftsResp);
  const actors =
    selectSuccessData<{ actors: CredentialActorForRoleLookup[] }>(credentialsResp)?.actors ??
    EMPTY_ACTORS;

  const [selectedShift, setSelectedShift] = useState<{
    id: string;
    cashierId?: string;
    drawerId?: string;
  } | null>(null);
  const [billValues, setBillValues] = useState<Record<number, string>>({});
  const [coinsMinor, setCoinsMinor] = useState(0);
  const [otherMinor, setOtherMinor] = useState(0);
  const [error, setError] = useState('');

  const countedMinor = useMemo(
    () => computeDenominationTotal(billValues, coinsMinor + otherMinor),
    [billValues, coinsMinor, otherMinor],
  );

  // Item 4 fix: the safe is looked up by containerType (there is exactly one
  // safe per store), but the drawer MUST be looked up by the selected
  // shift's own drawerId — never "first drawer found" — since more than one
  // open shift means more than one drawer balance exists.
  const safeBalance = useMemo(() => findSafeBalance(balances), [balances]);
  const drawerBalance = useMemo(
    () => findContainerBalance(balances, selectedShift?.drawerId),
    [balances, selectedShift?.drawerId],
  );
  const role = useMemo(
    () => findActorRole(actors, selectedShift?.cashierId),
    [actors, selectedShift?.cashierId],
  );

  const actorId = selectedShift?.cashierId ?? '';
  const approvedById = session?.actorId ?? '';

  const mutation = useMutation({
    mutationFn: async () => {
      return createCashMovement(
        storeId,
        {
          type: 'change_fund',
          fromContainerType: 'safe',
          fromContainerId: safeBalance!.containerId,
          toContainerType: 'drawer',
          toContainerId: selectedShift?.drawerId ?? 'drawer-1',
          amountMinor: countedMinor,
          actorId,
          approvedById,
          reason: 'change_fund',
        },
        { headers: createIdempotencyHeaders() },
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: getListCashBalancesQueryKey(storeId) });
      navigate('/dashboard', { replace: true });
    },
    onError: (err: Error) => setError(err?.message ?? t('common.unexpectedError')),
  });

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      setError('');

      if (!selectedShift) {
        setError(t('cash.selectCashier'));
        return;
      }
      if (!selectedShift.drawerId) {
        setError(t('cash.selectShift'));
        return;
      }
      if (!safeBalance) {
        setError(t('common.loading'));
        return;
      }
      if (!countedMinor || countedMinor <= 0) {
        setError(t('validation.mustBePositive', { field: t('cash.countedAmount') }));
        return;
      }
      if (!actorId || !approvedById) {
        setError(t('cash.actorSelfApproval'));
        return;
      }
      if (!actorsMustDiffer(actorId, approvedById)) {
        setError(t('cash.actorSelfApproval'));
        return;
      }

      mutation.mutate();
    },
    [selectedShift, safeBalance, countedMinor, actorId, approvedById, mutation, t],
  );

  return (
    <div className="sr-terminal-shell">
      <TopBar onHandover={onHandover} onLock={onLock} />

      <main className="sr-terminal-main">
        <h1 className="sr-page-title">{t('cash.changeFundTitle')}</h1>
        <p className="sr-cash-op-intro">{t('cash.pageIntro.changeFund')}</p>

        <form onSubmit={handleSubmit} className="sr-cash-op-form">
          <CashierSelectModal
            shifts={shiftsData?.shifts ?? []}
            onSelect={setSelectedShift}
            triggerLabel={selectedShift ? selectedShift.cashierId : undefined}
          />

          {selectedShift && (
            <CashierIdentityBanner
              icon={<DownArrowIcon />}
              accent="green"
              directionLabel={t('cash.identity.directionChangeFund')}
              cashierId={selectedShift.cashierId ?? ''}
              role={role}
            />
          )}

          <div className="sr-cash-op-body">
            <div className="sr-cash-op-main">
              <DenominationGrid
                variant="issue"
                billValues={billValues}
                onBillValuesChange={setBillValues}
                coinsMinor={coinsMinor}
                onCoinsMinorChange={setCoinsMinor}
                otherMinor={otherMinor}
                onOtherMinorChange={setOtherMinor}
              />
            </div>

            <div className="sr-cash-op-side">
              <BeforeAfterPanel
                totalLabel={t('cash.beforeAfter.totalIssue')}
                totalMinor={countedMinor}
                containers={[
                  { kind: 'safe', beforeMinor: safeBalance?.balanceMinor, direction: 'decrease' },
                  {
                    kind: 'drawer',
                    beforeMinor: drawerBalance?.balanceMinor,
                    direction: 'increase',
                  },
                ]}
              />
              <OperationChecklist
                steps={[
                  t('cash.checklist.changeFund1'),
                  t('cash.checklist.changeFund2'),
                  t('cash.checklist.changeFund3'),
                ]}
              />
            </div>
          </div>

          <p className="muted">{t('cash.noExpectedAmountNote')}</p>

          {error && <p className="sr-field-error">{error}</p>}

          <div className="sr-button-row">
            <Button type="button" variant="ghost" onClick={() => navigate('/dashboard')}>
              {t('common.cancel')}
            </Button>
            <Button type="submit" disabled={mutation.isPending || !safeBalance}>
              {mutation.isPending
                ? t('common.submitting')
                : !safeBalance
                  ? t('common.loading')
                  : t('common.confirm')}
            </Button>
          </div>
        </form>
      </main>
    </div>
  );
}
