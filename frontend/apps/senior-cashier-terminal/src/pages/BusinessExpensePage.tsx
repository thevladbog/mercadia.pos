import { useMemo, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button, Input, Field, Label } from '@mercadia/ui';
import {
  useListCashBalances,
  createBusinessExpense,
  getListCashBalancesQueryKey,
} from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { getStoreId } from '@/api-client-config.js';
import {
  actorsMustDiffer,
  computeDenominationTotal,
  createIdempotencyHeaders,
  selectSuccessData,
} from '@/lib/cash-utils.js';
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
 * Redesigned Business Expense page (plan 022, Phase 5). Same treatment as
 * `BankCollectionPage` — `DenominationGrid` + `BeforeAfterPanel` (safe-only)
 * + `OperationChecklist` — with `actorId`/`approvedById` staying manual
 * free text, UNCHANGED (plan item 5): no shift/cashier context to derive
 * from here either.
 *
 * CodeRabbit fix (PR #86): submission is now blocked until the real safe
 * balance resolves — same guard `BankCollectionPage`/`ReceiveCashPage`/
 * `IssueChangeFundPage` now have for the same
 * `safeBalance?.containerId ?? 'safe-1'` placeholder-fallback risk.
 */
export function BusinessExpensePage() {
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

  const [recipient, setRecipient] = useState('');
  const [reason, setReason] = useState('');
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
      return createBusinessExpense(
        storeId,
        {
          safeId: safeBalance!.containerId,
          payeeId: recipient,
          amountMinor: countedMinor,
          reason,
          actorId,
          approvedById,
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

      if (!recipient) {
        setError(t('validation.required', { field: t('cash.expenseRecipient') }));
        return;
      }
      if (!reason) {
        setError(t('validation.required', { field: t('cash.expenseReason') }));
        return;
      }
      if (!countedMinor || countedMinor <= 0) {
        setError(t('validation.mustBePositive', { field: t('cash.countedAmount') }));
        return;
      }
      if (!safeBalance) {
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

      mutation.mutate();
    },
    [recipient, reason, countedMinor, safeBalance, actorId, approvedById, mutation, t],
  );

  return (
    <div className="sr-terminal-shell">
      <TopBar onHandover={onHandover} onLock={onLock} />

      <main className="sr-terminal-main">
        <h1 className="sr-page-title">{t('cash.expenseTitle')}</h1>

        <form onSubmit={handleSubmit} className="sr-cash-op-form">
          <div className="sr-cash-op-fields">
            <Field>
              <Label>{t('cash.expenseRecipient')}</Label>
              <Input value={recipient} onChange={(e) => setRecipient(e.target.value)} required />
            </Field>

            <Field>
              <Label>{t('cash.expenseReason')}</Label>
              <Input
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                placeholder={t('cash.expenseReasonPlaceholder')}
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
                totalLabel={t('cash.beforeAfter.totalExpense')}
                totalMinor={countedMinor}
                containers={[
                  { kind: 'safe', beforeMinor: safeBalance?.balanceMinor, direction: 'decrease' },
                ]}
              />
              <OperationChecklist
                steps={[
                  t('cash.checklist.expense1'),
                  t('cash.checklist.expense2'),
                  t('cash.checklist.expense3'),
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
