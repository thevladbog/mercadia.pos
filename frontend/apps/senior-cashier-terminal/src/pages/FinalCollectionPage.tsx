import { useMemo, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button, Input, Field, Label } from '@mercadia/ui';
import {
  useListCashBalances,
  useListOpenStoreShifts,
  useGetCredentialManagement,
  closeShift,
  getListOpenStoreShiftsQueryKey,
  getListCashBalancesQueryKey,
} from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { useAuth } from '@/auth/AuthProvider.js';
import { getStoreId } from '@/api-client-config.js';
import { actorsMustDiffer, computeDenominationTotal, selectSuccessData } from '@/lib/cash-utils.js';
import { useTopBarActions } from '@/lib/use-topbar-actions.js';
import { CashierSelectModal } from '@/components/CashierSelectModal.js';
import { MismatchDialog } from '@/components/MismatchDialog.js';
import { TopBar } from '@/components/TopBar.js';

import { BoxIcon } from './dashboard/icons.js';
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
 * Redesigned Final Collection page (plan 022, Phase 5; design screen 03c's
 * recount half only — the critical-operations review panel and
 * two-signature close are OUT of scope, see plan item 2). Same shape as
 * `IssueChangeFundPage`/`ReceiveCashPage`, `MismatchDialog` reused
 * unchanged.
 *
 * Item 3 bug fix: `expectedMinor` used to be sourced from
 * `selectedShift?.closingCashMinor ?? 0` — `Shift.ClosingCashMinor` is set
 * ONLY inside `Shift.Close()` and is the Go zero-value (`0`) for every OPEN
 * shift, so the mismatch check was silently comparing the operator's real
 * count against a hardcoded `0` (it fired unless the operator counted
 * exactly 0). Fixed here: `expectedMinor` now comes from the real current
 * drawer balance via `useListCashBalances`, matched by the selected shift's
 * own `drawerId` (never "first drawer found" — item 4). Submission is
 * disabled until that balances query has resolved, so the mismatch check
 * never runs against a transient "not loaded yet" value.
 *
 * `actorId`/`approvedById` are auto-derived (plan 022 item 5), same as the
 * other 2 shift-based pages.
 */
export function FinalCollectionPage() {
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
  const shiftsData = useMemo(
    () =>
      selectSuccessData<{ shifts: { id: string; cashierId: string; drawerId: string }[] }>(
        shiftsResp,
      ),
    [shiftsResp],
  );
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
  const [closingCashInput, setClosingCashInput] = useState('');
  const [safeId, setSafeId] = useState('');
  const [error, setError] = useState('');
  const [showMismatch, setShowMismatch] = useState(false);

  const countedMinor = useMemo(
    () => computeDenominationTotal(billValues, coinsMinor + otherMinor),
    [billValues, coinsMinor, otherMinor],
  );
  const closingCashMinor = useMemo(
    () => Math.round(parseFloat(closingCashInput || '0') * 100),
    [closingCashInput],
  );

  const safeBalance = useMemo(() => findSafeBalance(balances), [balances]);
  const drawerBalance = useMemo(
    () => findContainerBalance(balances, selectedShift?.drawerId),
    [balances, selectedShift?.drawerId],
  );
  const role = useMemo(
    () => findActorRole(actors, selectedShift?.cashierId),
    [actors, selectedShift?.cashierId],
  );

  const balancesLoaded = balancesResp !== undefined;
  const expectedMinor = drawerBalance?.balanceMinor ?? 0;

  const actorId = selectedShift?.cashierId ?? '';
  const approvedById = session?.actorId ?? '';

  const mutation = useMutation({
    mutationFn: async () => {
      return closeShift(selectedShift!.id, {
        closingCashMinor,
        safeId: safeId || undefined,
        actorId,
        approvedById,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: getListOpenStoreShiftsQueryKey(storeId) });
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
        setError(t('cash.selectShift'));
        return;
      }
      if (!balancesLoaded) {
        setError(t('common.loading'));
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

      if (countedMinor !== expectedMinor) {
        setShowMismatch(true);
        return;
      }

      mutation.mutate();
    },
    [
      selectedShift,
      balancesLoaded,
      actorId,
      approvedById,
      countedMinor,
      expectedMinor,
      mutation,
      t,
    ],
  );

  const handleResolveMismatch = useCallback(() => {
    setShowMismatch(false);
    mutation.mutate();
  }, [mutation]);

  return (
    <div className="sr-terminal-shell">
      <TopBar onHandover={onHandover} onLock={onLock} />

      <main className="sr-terminal-main">
        <h1 className="sr-page-title">{t('cash.finalCollectionTitle')}</h1>
        <p className="sr-cash-op-intro">{t('cash.pageIntro.finalCollection')}</p>

        <form onSubmit={handleSubmit} className="sr-cash-op-form">
          <CashierSelectModal
            shifts={shiftsData?.shifts ?? []}
            onSelect={setSelectedShift}
            triggerLabel={selectedShift ? selectedShift.cashierId : undefined}
          />

          {selectedShift && (
            <CashierIdentityBanner
              icon={<BoxIcon />}
              accent="red"
              directionLabel={t('cash.identity.directionFinalCollection')}
              cashierId={selectedShift.cashierId ?? ''}
              role={role}
            />
          )}

          <div className="sr-cash-op-body">
            <div className="sr-cash-op-main">
              <DenominationGrid
                variant="recount"
                billValues={billValues}
                onBillValuesChange={setBillValues}
                coinsMinor={coinsMinor}
                onCoinsMinorChange={setCoinsMinor}
                otherMinor={otherMinor}
                onOtherMinorChange={setOtherMinor}
              />

              <Field className="sr-cash-op-fields">
                <Label>{t('cash.closingCash')}</Label>
                <Input
                  type="number"
                  min={0}
                  step={0.01}
                  value={closingCashInput}
                  onChange={(e) => setClosingCashInput(e.target.value)}
                  placeholder="0.00"
                />
              </Field>

              <Field className="sr-cash-op-fields">
                <Label>{t('cash.sourceSafe')}</Label>
                <Input
                  value={safeId}
                  onChange={(e) => setSafeId(e.target.value)}
                  placeholder="safe-1"
                />
              </Field>
            </div>

            <div className="sr-cash-op-side">
              <BeforeAfterPanel
                totalLabel={t('cash.beforeAfter.totalFinalCollection')}
                totalMinor={countedMinor}
                containers={[
                  {
                    kind: 'drawer',
                    beforeMinor: drawerBalance?.balanceMinor,
                    direction: 'decrease',
                  },
                  { kind: 'safe', beforeMinor: safeBalance?.balanceMinor, direction: 'increase' },
                ]}
              />
              <OperationChecklist
                steps={[
                  t('cash.checklist.finalCollection1'),
                  t('cash.checklist.finalCollection2'),
                  t('cash.checklist.finalCollection3'),
                ]}
              />
            </div>
          </div>

          {error && <p className="sr-field-error">{error}</p>}

          <div className="sr-button-row">
            <Button type="button" variant="ghost" onClick={() => navigate('/dashboard')}>
              {t('common.cancel')}
            </Button>
            <Button type="submit" disabled={mutation.isPending || !balancesLoaded}>
              {mutation.isPending
                ? t('common.submitting')
                : !balancesLoaded
                  ? t('common.loading')
                  : t('common.confirm')}
            </Button>
          </div>
        </form>

        <MismatchDialog
          expectedMinor={expectedMinor}
          countedMinor={countedMinor}
          open={showMismatch}
          onClose={() => setShowMismatch(false)}
          onResolve={handleResolveMismatch}
        />
      </main>
    </div>
  );
}
