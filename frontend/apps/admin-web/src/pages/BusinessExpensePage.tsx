import { useMemo, useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { Button } from '@mercadia/ui';
import { createCashMovement, useListCashBalances } from '@mercadia/api-clients-store-edge';
import { useListStores } from '@mercadia/api-clients-central';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import {
  DenominationInput,
  computeDenominationTotal,
} from '@/components/senior-cashier/DenominationInput.js';
import { StorePicker } from '@/components/StorePicker.js';
import { readStoreFromSearchParams } from '@/pages/store-routes.js';
import { createIdempotencyHeaders, invalidateSafeQueries } from '@/pages/cash-mutation-utils.js';

export function BusinessExpensePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchParams] = useSearchParams();
  const initialStoreId = readStoreFromSearchParams(searchParams);
  const [selectedStoreId, setSelectedStoreId_raw] = useState<string | null>(initialStoreId);
  const setSelectedStoreId = (id: string | null) => {
    setSelectedStoreId_raw(id);
    setFromContainerId('');
    setDenominations({});
    setActorId('');
    setApprovedById('');
    setErrorMessage(null);
  };

  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const activeStoreId = selectedStoreId ?? stores[0]?.id ?? '';

  const balancesQuery = useListCashBalances(activeStoreId, {
    query: { enabled: activeStoreId.length > 0 },
  });
  const balances = balancesQuery.data?.status === 200 ? balancesQuery.data.data.balances : null;

  const safeContainers = useMemo(() => {
    if (!balances) return [];
    return balances.filter((b) => b.containerType === 'safe');
  }, [balances]);

  const [fromContainerId, setFromContainerId] = useState('');
  const [denominations, setDenominations] = useState<Record<number, string>>({});
  const [expenseReason, setExpenseReason] = useState('');
  const [actorId, setActorId] = useState('');
  const [approvedById, setApprovedById] = useState('');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: async () => {
      const totalMinor = computeDenominationTotal(denominations);
      if (totalMinor <= 0) throw new Error(t('seniorCashier.amountRequired'));
      if (!expenseReason.trim()) throw new Error(t('seniorCashier.expenseReasonRequired'));
      if (!fromContainerId) throw new Error(t('seniorCashier.selectSafe'));
      return createCashMovement(
        activeStoreId,
        {
          type: 'expense',
          amountMinor: totalMinor,
          fromContainerId,
          fromContainerType: 'safe',
          toContainerId: fromContainerId,
          toContainerType: 'safe',
          actorId: actorId.trim(),
          approvedById: approvedById.trim(),
          reason: expenseReason.trim(),
        },
        { headers: createIdempotencyHeaders() },
      );
    },
    onSuccess: async (response) => {
      if (response.status === 202) {
        await invalidateSafeQueries(queryClient, activeStoreId);
        navigate(`/senior-cashier/dashboard?store=${encodeURIComponent(activeStoreId)}`, {
          state: { notice: t('seniorCashier.postedSuccessfully') },
        });
      }
    },
    onError: (error) => setErrorMessage(getApiErrorMessage(error)),
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage(null);
    if (!actorId.trim() || !approvedById.trim()) {
      setErrorMessage(t('seniorCashier.actorAndApproverRequired'));
      return;
    }
    if (actorId.trim() === approvedById.trim()) {
      setErrorMessage(t('safe.forms.validation.selfApproval'));
      return;
    }
    mutation.mutate();
  }

  if (!activeStoreId) {
    return (
      <div className="panel">
        <h1>{t('seniorCashier.businessExpense')}</h1>
        <p className="muted">{t('common.selectStore')}</p>
      </div>
    );
  }

  return (
    <div className="panel">
      <h1>{t('seniorCashier.businessExpense')}</h1>
      <StorePicker stores={stores} value={activeStoreId} onChange={setSelectedStoreId} />
      <section className="card">
        <form className="stack" onSubmit={handleSubmit}>
          <label>
            {t('seniorCashier.sourceSafe')}
            <select value={fromContainerId} onChange={(e) => setFromContainerId(e.target.value)}>
              <option value="">—</option>
              {safeContainers.map((c) => (
                <option key={c.containerId} value={c.containerId}>
                  {c.containerId}
                </option>
              ))}
            </select>
          </label>
          <label>
            {t('seniorCashier.reason')}
            <input
              type="text"
              value={expenseReason}
              onChange={(e) => setExpenseReason(e.target.value)}
            />
          </label>
          <fieldset>
            <legend>{t('seniorCashier.enterDenominations')}</legend>
            <DenominationInput values={denominations} onChange={setDenominations} />
          </fieldset>
          <label>
            {t('seniorCashier.confirmBySenior')}
            <input type="text" value={actorId} onChange={(e) => setActorId(e.target.value)} />
          </label>
          <label>
            {t('seniorCashier.confirmBySecondPerson')}
            <input
              type="text"
              value={approvedById}
              onChange={(e) => setApprovedById(e.target.value)}
            />
          </label>
          {errorMessage ? <p className="error">{errorMessage}</p> : null}
          <Button disabled={mutation.isPending} type="submit">
            {mutation.isPending ? t('common.submitting') : t('seniorCashier.confirmOperation')}
          </Button>
        </form>
      </section>
    </div>
  );
}
