import { useMemo, useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { Button, Input } from '@mercadia/ui';
import { closeShift, useListOpenStoreShifts } from '@mercadia/api-clients-store-edge';
import type { ListOpenStoreShifts200ShiftsItem } from '@mercadia/api-clients-store-edge';
import { useListStores } from '@mercadia/api-clients-central';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import {
  DenominationInput,
  computeDenominationTotal,
} from '@/components/senior-cashier/DenominationInput.js';
import { CashierSelectModal } from '@/components/senior-cashier/CashierSelectModal.js';
import { MismatchDialog } from '@/components/senior-cashier/MismatchDialog.js';
import { StorePicker } from '@/components/StorePicker.js';
import { readStoreFromSearchParams } from '@/pages/store-routes.js';
import { createIdempotencyHeaders, invalidateSafeQueries } from '@/pages/cash-mutation-utils.js';

export function ShiftHandoverPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchParams] = useSearchParams();
  const initialStoreId = readStoreFromSearchParams(searchParams);
  const [selectedStoreId, setSelectedStoreId_raw] = useState<string | null>(initialStoreId);
  const setSelectedStoreId = (id: string | null) => {
    setSelectedStoreId_raw(id);
    setSelectedShift(null);
    setIncomingCashierId('');
    setActorId('');
    setApprovedById('');
    setDenominations({});
    setErrorMessage(null);
  };

  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const activeStoreId = selectedStoreId ?? stores[0]?.id ?? '';

  const shiftsQuery = useListOpenStoreShifts(activeStoreId, {
    query: { enabled: activeStoreId.length > 0 },
  });
  const shifts = shiftsQuery.data?.status === 200 ? shiftsQuery.data.data.shifts : null;

  const openShifts = useMemo(() => {
    if (!shifts) return [];
    return shifts;
  }, [shifts]);

  const [selectedShift, setSelectedShift] = useState<ListOpenStoreShifts200ShiftsItem | null>(null);
  const [incomingCashierId, setIncomingCashierId] = useState('');
  const [actorId, setActorId] = useState('');
  const [approvedById, setApprovedById] = useState('');
  const [denominations, setDenominations] = useState<Record<number, string>>({});
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [showMismatch, setShowMismatch] = useState(false);

  const countedMinor = computeDenominationTotal(denominations);
  const expectedMinor = selectedShift?.closingCashMinor ?? 0;
  const hasMismatch = selectedShift && expectedMinor !== countedMinor;

  const mutation = useMutation({
    mutationFn: async () => {
      if (!selectedShift) throw new Error(t('seniorCashier.selectShift'));
      if (!incomingCashierId.trim()) throw new Error(t('seniorCashier.incomingCashierRequired'));
      return closeShift(
        selectedShift.id,
        {
          closingCashMinor: countedMinor,
          actorId: actorId.trim(),
          approvedById: approvedById.trim(),
        },
        { headers: createIdempotencyHeaders() },
      );
    },
    onSuccess: async (response) => {
      if (response.status === 202) {
        await invalidateSafeQueries(queryClient, activeStoreId);
        navigate(`/senior-cashier/dashboard?store=${encodeURIComponent(activeStoreId)}`, {
          state: { notice: t('seniorCashier.shiftHandoverReady') },
        });
      }
    },
    onError: (error) => setErrorMessage(getApiErrorMessage(error)),
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage(null);
    if (!selectedShift) {
      setErrorMessage(t('seniorCashier.selectShift'));
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
    if (hasMismatch) {
      setShowMismatch(true);
      return;
    }
    mutation.mutate();
  }

  if (!activeStoreId) {
    return (
      <div className="panel">
        <h1>{t('seniorCashier.shiftHandover')}</h1>
        <p className="muted">{t('common.selectStore')}</p>
      </div>
    );
  }

  return (
    <div className="panel">
      <h1>{t('seniorCashier.shiftHandover')}</h1>
      <StorePicker stores={stores} value={activeStoreId} onChange={setSelectedStoreId} />
      <section className="card">
        <form className="stack" onSubmit={handleSubmit}>
          <CashierSelectModal
            shifts={openShifts}
            onSelect={(shift) => setSelectedShift(shift)}
            triggerLabel={t('seniorCashier.selectOutgoingCashier')}
          />
          {selectedShift ? (
            <p>
              {t('seniorCashier.expectedAmount')}: {(expectedMinor / 100).toFixed(2)} ₽
            </p>
          ) : null}
          <label>
            {t('seniorCashier.incomingCashier')}
            {t('seniorCashier.idSuffix')}
            <Input
              type="text"
              value={incomingCashierId}
              onChange={(e) => setIncomingCashierId(e.target.value)}
            />
          </label>
          <fieldset>
            <legend>{t('seniorCashier.enterDenominations')}</legend>
            <DenominationInput values={denominations} onChange={setDenominations} />
          </fieldset>
          <label>
            {t('seniorCashier.confirmBySenior')}
            <Input type="text" value={actorId} onChange={(e) => setActorId(e.target.value)} />
          </label>
          <label>
            {t('seniorCashier.confirmBySecondPerson')}
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
      {selectedShift ? (
        <MismatchDialog
          expectedMinor={expectedMinor}
          countedMinor={countedMinor}
          open={showMismatch}
          onClose={() => setShowMismatch(false)}
          onResolve={() => {
            setShowMismatch(false);
            mutation.mutate();
          }}
        />
      ) : null}
    </div>
  );
}
