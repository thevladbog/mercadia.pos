import type { ListCashBalances200BalancesItem } from '@mercadia/api-clients-store-edge';
import { createCashRecount } from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { ActorFields } from '@/components/cash/ActorFields.js';
import { AmountField } from '@/components/cash/AmountField.js';
import { CashModal } from '@/components/cash/CashModal.js';
import { ContainerSelect } from '@/components/cash/ContainerSelect.js';
import { firstContainerByType } from '@/pages/cash-container-utils.js';
import {
  actorsMustDiffer,
  createIdempotencyHeaders,
  invalidateSafeQueries,
  parseRublesToMinor,
} from '@/pages/cash-mutation-utils.js';
import { formatMinorAmount } from '@/pages/reporting-utils.js';
import { storePageHref } from '@/pages/store-routes.js';

type CreateRecountModalProps = {
  storeId: string;
  balances: ListCashBalances200BalancesItem[];
  onClose: () => void;
};

export function CreateRecountModal({ storeId, balances, onClose }: CreateRecountModalProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const defaultContainer = firstContainerByType(balances, 'safe') ?? balances[0];

  const [containerId, setContainerId] = useState(defaultContainer?.containerId ?? '');
  const [countedRub, setCountedRub] = useState('');
  const [reason, setReason] = useState('');
  const [actorId, setActorId] = useState('');
  const [approvedById, setApprovedById] = useState('');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const selectedContainer = balances.find((b) => b.containerId === containerId);

  const mutation = useMutation({
    mutationFn: async () => {
      if (!selectedContainer) {
        throw new Error('missing container');
      }
      const countedMinor = parseRublesToMinor(countedRub);
      if (countedMinor == null) {
        throw new Error('invalid amount');
      }
      return createCashRecount(
        storeId,
        {
          containerId: selectedContainer.containerId,
          containerType: selectedContainer.containerType,
          countedMinor,
          actorId: actorId.trim(),
          ...(approvedById.trim() ? { approvedById: approvedById.trim() } : {}),
          ...(reason.trim() ? { reason: reason.trim() } : {}),
        },
        { headers: createIdempotencyHeaders() },
      );
    },
    onSuccess: async (response) => {
      if (response.status === 202) {
        await invalidateSafeQueries(queryClient, storeId);
        void navigate(storePageHref('/store/safe', storeId), {
          state: { notice: t('safe.notices.recountSuccess') },
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

    const countedMinor = parseRublesToMinor(countedRub);
    if (countedMinor == null || !selectedContainer) {
      setErrorMessage(t('safe.forms.validation.amount'));
      return;
    }

    const hasDiscrepancy = countedMinor !== selectedContainer.balanceMinor;
    if (hasDiscrepancy && !actorsMustDiffer(actorId, approvedById)) {
      setErrorMessage(t('safe.forms.validation.selfApproval'));
      return;
    }

    mutation.mutate();
  }

  const expectedLabel = selectedContainer
    ? formatMinorAmount(selectedContainer.balanceMinor)
    : t('common.emDash');

  return (
    <CashModal
      errorMessage={errorMessage}
      isSubmitting={mutation.isPending}
      submitLabel={t('safe.forms.submit')}
      title={t('safe.actions.recountSafe')}
      onClose={onClose}
      onSubmit={handleSubmit}
    >
      <ContainerSelect
        containers={balances}
        label={t('safe.forms.container')}
        value={containerId}
        onChange={setContainerId}
      />
      <p className="muted">
        {t('safe.forms.expectedBalance')}: {expectedLabel}
      </p>
      <AmountField value={countedRub} onChange={setCountedRub} />
      <label className="field">
        <span>{t('safe.forms.reason')}</span>
        <input value={reason} onChange={(event) => setReason(event.target.value)} />
      </label>
      <ActorFields
        actorId={actorId}
        approvedById={approvedById}
        requireApprover={false}
        onActorIdChange={setActorId}
        onApprovedByIdChange={setApprovedById}
      />
    </CashModal>
  );
}
