import { useMemo, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button, Input, Field, Label } from '@mercadia/ui';
import {
  useListCashBalances,
  createBankCollection,
  getListCashBalancesQueryKey,
} from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { getStoreId } from '@/api-client-config.js';
import { actorsMustDiffer, computeDenominationTotal, selectSuccessData } from '@/lib/cash-utils.js';
import { useTopBarActions } from '@/lib/use-topbar-actions.js';
import { TopBar } from '@/components/TopBar.js';

import { BeforeAfterPanel } from './cash-operations/BeforeAfterPanel.js';
import { DenominationGrid } from './cash-operations/DenominationGrid.js';
import { OperationChecklist } from './cash-operations/OperationChecklist.js';
import {
  findSafeBalance,
  type CashBalanceForLookup,
} from './cash-operations/cash-operations-data.js';
import './cash-operations/CashOperations.css';

// Stable empty-array fallback (module scope, never reassigned) — same
// convention as the other 4 cash-operation pages (see
// `IssueChangeFundPage.tsx`'s doc comment on the same constant).
const EMPTY_BALANCES: CashBalanceForLookup[] = [];

/**
 * Redesigned Bank Collection page (plan 022, Phase 5). Gets the same
 * `DenominationGrid` + `BeforeAfterPanel` (safe-only, no drawer pair) +
 * `OperationChecklist` treatment as the other 4 pages, but its
 * `actorId`/`approvedById` inputs stay manual free text, UNCHANGED (plan
 * item 5) — this page has no shift/cashier selection at all (it's a
 * safe-only operation against an external collection service), so there is
 * no session-derived value for either field.
 */
export function BankCollectionPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { onHandover, onLock } = useTopBarActions();
  const storeId = useMemo(() => getStoreId(), []);

  const { data: balancesResp } = useListCashBalances(storeId);
  const balances =
    selectSuccessData<{ balances: CashBalanceForLookup[] }>(balancesResp)?.balances ??
    EMPTY_BALANCES;
  const safeBalance = useMemo(() => findSafeBalance(balances), [balances]);

  const [collectorName, setCollectorName] = useState('');
  const [contractNumber, setContractNumber] = useState('');
  const [bagSealNumber, setBagSealNumber] = useState('');
  const [billValues, setBillValues] = useState<Record<number, string>>({});
  const [coinsMinor, setCoinsMinor] = useState(0);
  const [otherMinor, setOtherMinor] = useState(0);
  const [actorId, setActorId] = useState('');
  const [approvedById, setApprovedById] = useState('');
  const [error, setError] = useState('');

  const countedMinor = useMemo(
    () => computeDenominationTotal(billValues, coinsMinor + otherMinor),
    [billValues, coinsMinor, otherMinor],
  );

  const mutation = useMutation({
    mutationFn: async () => {
      return createBankCollection(storeId, {
        safeId: safeBalance?.containerId ?? 'safe-1',
        bankContainerId: contractNumber || 'bank-1',
        amountMinor: countedMinor || 1,
        actorId,
        approvedById,
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

      if (!collectorName) {
        setError(t('validation.required', { field: t('cash.collectorName') }));
        return;
      }
      if (!contractNumber) {
        setError(t('validation.required', { field: t('cash.collectorContract') }));
        return;
      }
      if (!bagSealNumber) {
        setError(t('validation.required', { field: t('cash.bagSealNumber') }));
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
    [
      collectorName,
      contractNumber,
      bagSealNumber,
      countedMinor,
      actorId,
      approvedById,
      mutation,
      t,
    ],
  );

  return (
    <div className="sr-terminal-shell">
      <TopBar onHandover={onHandover} onLock={onLock} />

      <main className="sr-terminal-main">
        <h1 className="sr-page-title">{t('cash.bankCollectionTitle')}</h1>

        <form onSubmit={handleSubmit} className="sr-cash-op-form">
          <p className="muted">{t('cash.collectorInfo')}</p>

          <div className="sr-cash-op-fields">
            <Field>
              <Label>{t('cash.collectorName')}</Label>
              <Input
                value={collectorName}
                onChange={(e) => setCollectorName(e.target.value)}
                required
              />
            </Field>

            <Field>
              <Label>{t('cash.collectorContract')}</Label>
              <Input
                value={contractNumber}
                onChange={(e) => setContractNumber(e.target.value)}
                required
              />
            </Field>

            <Field>
              <Label>{t('cash.bagSealNumber')}</Label>
              <Input
                value={bagSealNumber}
                onChange={(e) => setBagSealNumber(e.target.value)}
                required
              />
            </Field>
          </div>

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
                totalLabel={t('cash.beforeAfter.totalBankCollection')}
                totalMinor={countedMinor}
                containers={[
                  { kind: 'safe', beforeMinor: safeBalance?.balanceMinor, direction: 'decrease' },
                ]}
              />
              <OperationChecklist
                steps={[
                  t('cash.checklist.bankCollection1'),
                  t('cash.checklist.bankCollection2'),
                  t('cash.checklist.bankCollection3'),
                ]}
              />
            </div>
          </div>

          <p className="muted">{t('cash.confirmTwoPerson')}</p>

          <div className="sr-cash-op-fields">
            <Field>
              <Label>{t('cash.actorId')}</Label>
              <Input value={actorId} onChange={(e) => setActorId(e.target.value)} required />
            </Field>

            <Field>
              <Label>{t('cash.approvedById')}</Label>
              <Input
                value={approvedById}
                onChange={(e) => setApprovedById(e.target.value)}
                required
              />
            </Field>
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
      </main>
    </div>
  );
}
