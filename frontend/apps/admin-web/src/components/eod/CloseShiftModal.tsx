import type {
  ListCashBalances200BalancesItem,
  ListOpenStoreShifts200ShiftsItem,
} from '@mercadia/api-clients-store-edge';
import { closeShift } from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useMemo, useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { ActorFields } from '@/components/cash/ActorFields.js';
import { CashModal } from '@/components/cash/CashModal.js';
import { ContainerSelect } from '@/components/cash/ContainerSelect.js';
import { containersByType, firstContainerByType } from '@/pages/cash-container-utils.js';
import { actorsMustDiffer, createIdempotencyHeaders } from '@/pages/cash-mutation-utils.js';
import { formatMinorAmount } from '@/pages/reporting-utils.js';
import {
  invalidateShiftCloseQueries,
  parseNonNegativeRublesToMinor,
} from '@/pages/shift-mutation-utils.js';
import { storePageHref } from '@/pages/store-routes.js';

type CloseShiftModalProps = {
  storeId: string;
  shift: ListOpenStoreShifts200ShiftsItem;
  balances: ListCashBalances200BalancesItem[];
  onClose: () => void;
};

export function CloseShiftModal({ storeId, shift, balances, onClose }: CloseShiftModalProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const safeContainers = useMemo(() => containersByType(balances, 'safe'), [balances]);
  const defaultSafe = useMemo(() => firstContainerByType(balances, 'safe'), [balances]);

  const [closingCashRub, setClosingCashRub] = useState('0.00');
  const [safeId, setSafeId] = useState(defaultSafe?.containerId ?? '');
  const effectiveSafeId = safeId || defaultSafe?.containerId || '';
  const [actorId, setActorId] = useState('');
  const [approvedById, setApprovedById] = useState('');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const closingCashMinor = parseNonNegativeRublesToMinor(closingCashRub);
  const requiresCollection = closingCashMinor != null && closingCashMinor > 0;

  const mutation = useMutation({
    mutationFn: async () => {
      if (closingCashMinor == null) {
        throw new Error('invalid amount');
      }
      return closeShift(
        shift.id,
        {
          closingCashMinor,
          ...(requiresCollection
            ? {
                safeId: effectiveSafeId,
                actorId: actorId.trim(),
                approvedById: approvedById.trim(),
              }
            : {}),
        },
        { headers: createIdempotencyHeaders() },
      );
    },
    onSuccess: async (response) => {
      if (response.status === 202) {
        await invalidateShiftCloseQueries(queryClient, storeId, shift.operationalDayId);
        void navigate(storePageHref('/store/eod', storeId), {
          state: { notice: t('eod.notices.closeShiftSuccess') },
        });
        onClose();
        return;
      }
      setErrorMessage(t('common.unexpectedError'));
    },
    onError: (error) => {
      setErrorMessage(getApiErrorMessage(error));
    },
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage(null);

    if (closingCashMinor == null) {
      setErrorMessage(t('eod.forms.validation.closingCashRequired'));
      return;
    }

    if (requiresCollection) {
      if (effectiveSafeId.length === 0) {
        setErrorMessage(t('eod.forms.validation.safeRequired'));
        return;
      }
      if (actorId.trim().length === 0 || approvedById.trim().length === 0) {
        setErrorMessage(t('eod.forms.validation.actorsRequired'));
        return;
      }
      if (!actorsMustDiffer(actorId, approvedById)) {
        setErrorMessage(t('safe.forms.validation.selfApproval'));
        return;
      }
    }

    mutation.mutate();
  }

  return (
    <CashModal
      errorMessage={errorMessage}
      isSubmitting={mutation.isPending}
      submitLabel={t('eod.actions.closeShift')}
      title={t('eod.forms.closeShiftTitle')}
      onClose={onClose}
      onSubmit={handleSubmit}
    >
      <p className="muted">{t('eod.forms.closeShiftBody')}</p>

      <dl className="kpi-grid">
        <div>
          <dt>{t('monitoring.cashier')}</dt>
          <dd>{shift.cashierId}</dd>
        </div>
        <div>
          <dt>{t('eod.terminalId')}</dt>
          <dd>{shift.terminalId}</dd>
        </div>
        <div>
          <dt>{t('safe.container')}</dt>
          <dd>{shift.drawerId}</dd>
        </div>
        <div>
          <dt>{t('eod.openingCash')}</dt>
          <dd>{formatMinorAmount(shift.openingCashMinor)}</dd>
        </div>
      </dl>

      <label className="field">
        <span>{t('eod.forms.closingCash')}</span>
        <input
          inputMode="decimal"
          min="0"
          required
          step="0.01"
          type="number"
          value={closingCashRub}
          onChange={(event) => setClosingCashRub(event.target.value)}
        />
      </label>

      {requiresCollection ? (
        <>
          <ContainerSelect
            containers={safeContainers}
            label={t('safe.forms.safeContainer')}
            value={effectiveSafeId}
            onChange={setSafeId}
          />
          <ActorFields
            actorId={actorId}
            approvedById={approvedById}
            onActorIdChange={setActorId}
            onApprovedByIdChange={setApprovedById}
          />
        </>
      ) : null}
    </CashModal>
  );
}
