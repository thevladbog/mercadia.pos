import { useMemo, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button, Input, Field, Label } from '@mercadia/ui';
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
import { actorsMustDiffer, computeDenominationTotal, selectSuccessData } from '@/lib/cash-utils.js';
import { useTopBarActions } from '@/lib/use-topbar-actions.js';
import { CashierSelectModal } from '@/components/CashierSelectModal.js';
import { MismatchDialog } from '@/components/MismatchDialog.js';
import { TopBar } from '@/components/TopBar.js';

import { UpArrowIcon } from './dashboard/icons.js';
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
 * Redesigned Receive Cash page (plan 022, Phase 5; design screen 03b). Same
 * shape as `IssueChangeFundPage`, plus the existing optional "expected
 * amount" input and `MismatchDialog`, both reused unchanged.
 *
 * `actorId`/`approvedById` are auto-derived (plan 022 item 5), same as
 * `IssueChangeFundPage` — see that file's doc comment.
 */
export function ReceiveCashPage() {
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
  const [expectedInput, setExpectedInput] = useState('');
  const [billValues, setBillValues] = useState<Record<number, string>>({});
  const [coinsMinor, setCoinsMinor] = useState(0);
  const [otherMinor, setOtherMinor] = useState(0);
  const [error, setError] = useState('');
  const [showMismatch, setShowMismatch] = useState(false);

  const expectedMinor = useMemo(
    () => Math.round(parseFloat(expectedInput || '0') * 100),
    [expectedInput],
  );
  const countedMinor = useMemo(
    () => computeDenominationTotal(billValues, coinsMinor + otherMinor),
    [billValues, coinsMinor, otherMinor],
  );

  // Item 4 fix: drawer balance must come from the selected shift's own
  // drawerId, never "first drawer found" (see IssueChangeFundPage).
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
      return createCashMovement(storeId, {
        type: 'cash_out',
        fromContainerType: 'drawer',
        fromContainerId: selectedShift?.drawerId ?? 'drawer-1',
        toContainerType: 'safe',
        toContainerId: safeBalance?.containerId ?? 'safe-1',
        amountMinor: countedMinor || 1,
        actorId,
        approvedById,
        reason: 'revenue_collection',
      });
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
      if (!actorId || !approvedById) {
        setError(t('cash.actorSelfApproval'));
        return;
      }
      if (!actorsMustDiffer(actorId, approvedById)) {
        setError(t('cash.actorSelfApproval'));
        return;
      }

      if (expectedMinor > 0 && expectedMinor !== countedMinor) {
        setShowMismatch(true);
        return;
      }

      mutation.mutate();
    },
    [selectedShift, actorId, approvedById, expectedMinor, countedMinor, mutation, t],
  );

  const handleResolveMismatch = useCallback(() => {
    setShowMismatch(false);
    mutation.mutate();
  }, [mutation]);

  return (
    <div className="sr-terminal-shell">
      <TopBar onHandover={onHandover} onLock={onLock} />

      <main className="sr-terminal-main">
        <h1 className="sr-page-title">{t('cash.receiveCashTitle')}</h1>
        <p className="sr-cash-op-intro">{t('cash.pageIntro.receiveCash')}</p>

        <form onSubmit={handleSubmit} className="sr-cash-op-form">
          <CashierSelectModal
            shifts={shiftsData?.shifts ?? []}
            onSelect={setSelectedShift}
            triggerLabel={selectedShift ? selectedShift.cashierId : undefined}
          />

          {selectedShift && (
            <CashierIdentityBanner
              icon={<UpArrowIcon />}
              accent="blue"
              directionLabel={t('cash.identity.directionReceiveCash')}
              cashierId={selectedShift.cashierId ?? ''}
              role={role}
            />
          )}

          <Field className="sr-cash-op-fields">
            <Label>{t('cash.expectedAmount')}</Label>
            <Input
              type="number"
              min={0}
              step={0.01}
              value={expectedInput}
              onChange={(e) => setExpectedInput(e.target.value)}
              placeholder="0.00"
            />
          </Field>

          <div className="sr-cash-op-body">
            <div className="sr-cash-op-main">
              <DenominationGrid
                variant="receive"
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
                totalLabel={t('cash.beforeAfter.totalReceive')}
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
                  t('cash.checklist.receiveCash1'),
                  t('cash.checklist.receiveCash2'),
                  t('cash.checklist.receiveCash3'),
                ]}
              />
            </div>
          </div>

          {error && <p className="sr-field-error">{error}</p>}

          <div className="sr-button-row">
            <Button type="button" variant="ghost" onClick={() => navigate('/dashboard')}>
              {t('common.cancel')}
            </Button>
            <Button type="submit" disabled={mutation.isPending}>
              {mutation.isPending ? t('common.submitting') : t('common.confirm')}
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
