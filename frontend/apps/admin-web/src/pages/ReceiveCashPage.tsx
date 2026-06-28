import { useMemo, useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { Button, Input } from '@mercadia/ui';
import {
  createCashMovement,
  useListCashBalances,
  useListOpenStoreShifts,
} from '@mercadia/api-clients-store-edge';
import { useListStores } from '@mercadia/api-clients-central';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import {
  DenominationInput,
  computeDenominationTotal,
} from '@/components/senior-cashier/DenominationInput.js';
import { CashierSelectModal } from '@/components/senior-cashier/CashierSelectModal.js';
import { StorePicker } from '@/components/StorePicker.js';
import { readStoreFromSearchParams } from '@/pages/store-routes.js';
import { createIdempotencyHeaders, invalidateSafeQueries } from '@/pages/cash-mutation-utils.js';

export function ReceiveCashPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchParams] = useSearchParams();
  const initialStoreId = readStoreFromSearchParams(searchParams);
  const [selectedStoreId, setSelectedStoreId_raw] = useState<string | null>(initialStoreId);
  const setSelectedStoreId = (id: string | null) => {
    setSelectedStoreId_raw(id);
    setSelectedDrawerId('');
    setSelectedSafeId('');
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
  const shiftsQuery = useListOpenStoreShifts(activeStoreId, {
    query: { enabled: activeStoreId.length > 0 },
  });

  const balances = balancesQuery.data?.status === 200 ? balancesQuery.data.data.balances : null;
  const shifts = shiftsQuery.data?.status === 200 ? shiftsQuery.data.data.shifts : null;

  const drawerContainers = useMemo(() => {
    if (!balances) return [];
    return balances.filter((b) => b.containerType === 'drawer');
  }, [balances]);

  const safeContainers = useMemo(() => {
    if (!balances) return [];
    return balances.filter((b) => b.containerType === 'safe');
  }, [balances]);

  const [selectedDrawerId, setSelectedDrawerId] = useState('');
  const [selectedSafeId, setSelectedSafeId] = useState('');
  const [denominations, setDenominations] = useState<Record<number, string>>({});
  const [expectedMinor, setExpectedMinor] = useState('');
  const [actorId, setActorId] = useState('');
  const [approvedById, setApprovedById] = useState('');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: async () => {
      const countedMinor = computeDenominationTotal(denominations);
      if (countedMinor <= 0) throw new Error(t('seniorCashier.countedAmountRequired'));
      return createCashMovement(
        activeStoreId,
        {
          type: 'cash_out',
          fromContainerId: selectedDrawerId,
          fromContainerType: 'drawer',
          toContainerId: selectedSafeId,
          toContainerType: 'safe',
          amountMinor: countedMinor,
          actorId: actorId.trim(),
          approvedById: approvedById.trim(),
          reason: 'revenue_collection',
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
    if (!selectedDrawerId || !selectedSafeId) {
      setErrorMessage(t('seniorCashier.selectSourceDrawerDestSafe'));
      return;
    }
    if (mismatch) {
      setErrorMessage(t('seniorCashier.mismatchDetected'));
      return;
    }
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
        <h1>{t('seniorCashier.receiveCash')}</h1>
        <p className="muted">{t('common.selectStore')}</p>
      </div>
    );
  }

  const countedMinor = computeDenominationTotal(denominations);
  const expectedInt = parseInt(expectedMinor || '0', 10);
  const mismatch = expectedInt > 0 && countedMinor !== expectedInt;

  return (
    <div className="panel">
      <h1>{t('seniorCashier.receiveCash')}</h1>
      <StorePicker stores={stores} value={activeStoreId} onChange={setSelectedStoreId} />
      <section className="card">
        <form className="stack" onSubmit={handleSubmit}>
          <label>
            {t('seniorCashier.cashDrawer')}
            <select value={selectedDrawerId} onChange={(e) => setSelectedDrawerId(e.target.value)}>
              <option value="">—</option>
              {drawerContainers.map((c) => (
                <option key={c.containerId} value={c.containerId}>
                  {c.containerId}
                </option>
              ))}
            </select>
          </label>
          {shifts && shifts.length > 0 ? (
            <CashierSelectModal
              shifts={shifts}
              onSelect={(shift) => setSelectedDrawerId(shift.drawerId)}
            />
          ) : null}
          <label>
            {t('seniorCashier.expectedAmount')}
            {t('seniorCashier.kopecksSuffix')}
            <Input
              type="number"
              value={expectedMinor}
              onChange={(e) => setExpectedMinor(e.target.value)}
            />
          </label>
          <fieldset>
            <legend>{t('seniorCashier.enterDenominations')}</legend>
            <DenominationInput values={denominations} onChange={setDenominations} />
          </fieldset>
          {mismatch ? (
            <div className="alert alert-warning">
              {t('seniorCashier.mismatchDetail', {
                expected: (expectedInt / 100).toFixed(2),
                counted: (countedMinor / 100).toFixed(2),
                diff: ((expectedInt - countedMinor) / 100).toFixed(2),
              })}
            </div>
          ) : null}
          <label>
            {t('seniorCashier.destinationSafe')}
            <select value={selectedSafeId} onChange={(e) => setSelectedSafeId(e.target.value)}>
              <option value="">—</option>
              {safeContainers.map((c) => (
                <option key={c.containerId} value={c.containerId}>
                  {c.containerId}
                </option>
              ))}
            </select>
          </label>
          <label>
            {t('seniorCashier.confirmBySenior')}
            {t('seniorCashier.idSuffix')}
            <Input type="text" value={actorId} onChange={(e) => setActorId(e.target.value)} />
          </label>
          <label>
            {t('seniorCashier.confirmBySecondPerson')}
            {t('seniorCashier.idSuffix')}
            <Input
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
