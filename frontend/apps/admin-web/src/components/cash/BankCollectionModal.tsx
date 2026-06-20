import type { ListCashBalances200BalancesItem } from '@mercadia/api-clients-store-edge';
import { createBankCollection } from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useMemo, useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { ActorFields } from '@/components/cash/ActorFields.js';
import { AmountField } from '@/components/cash/AmountField.js';
import { FormDialog } from '@mercadia/ui';
import { ContainerSelect } from '@/components/cash/ContainerSelect.js';
import { containersByType, firstContainerByType } from '@/pages/cash-container-utils.js';
import {
  actorsMustDiffer,
  createIdempotencyHeaders,
  invalidateSafeQueries,
  parseRublesToMinor,
} from '@/pages/cash-mutation-utils.js';
import { storePageHref } from '@/pages/store-routes.js';

type BankCollectionModalProps = {
  storeId: string;
  balances: ListCashBalances200BalancesItem[];
  onClose: () => void;
};

export function BankCollectionModal({ storeId, balances, onClose }: BankCollectionModalProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const safeContainers = useMemo(() => containersByType(balances, 'safe'), [balances]);
  const bankContainers = useMemo(() => containersByType(balances, 'bank'), [balances]);

  const defaultSafe = firstContainerByType(balances, 'safe');
  const defaultBank = firstContainerByType(balances, 'bank');

  const [safeId, setSafeId] = useState(defaultSafe?.containerId ?? '');
  const [bankContainerId, setBankContainerId] = useState(defaultBank?.containerId ?? '');
  const [amountRub, setAmountRub] = useState('');
  const [reason, setReason] = useState('');
  const [actorId, setActorId] = useState('');
  const [approvedById, setApprovedById] = useState('');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: async () => {
      const amountMinor = parseRublesToMinor(amountRub);
      if (amountMinor == null) {
        throw new Error('invalid amount');
      }
      return createBankCollection(
        storeId,
        {
          safeId,
          bankContainerId,
          amountMinor,
          actorId: actorId.trim(),
          approvedById: approvedById.trim(),
          ...(reason.trim() ? { reason: reason.trim() } : {}),
        },
        { headers: createIdempotencyHeaders() },
      );
    },
    onSuccess: async (response) => {
      if (response.status === 202) {
        await invalidateSafeQueries(queryClient, storeId);
        void navigate(storePageHref('/store/safe', storeId), {
          state: { notice: t('safe.notices.bankCollectionSuccess') },
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

    if (!actorsMustDiffer(actorId, approvedById)) {
      setErrorMessage(t('safe.forms.validation.selfApproval'));
      return;
    }
    if (parseRublesToMinor(amountRub) == null) {
      setErrorMessage(t('safe.forms.validation.amount'));
      return;
    }
    if (!safeId || !bankContainerId) {
      setErrorMessage(t('safe.forms.validation.containers'));
      return;
    }

    mutation.mutate();
  }

  return (
    <FormDialog
      cancelLabel={t('common.cancel')}
      errorMessage={errorMessage}
      isSubmitting={mutation.isPending}
      submitLabel={mutation.isPending ? t('common.submitting') : t('safe.forms.submit')}
      title={t('safe.actions.bankCollection')}
      onClose={onClose}
      onSubmit={handleSubmit}
    >
      <ContainerSelect
        containers={safeContainers}
        label={t('safe.forms.safeContainer')}
        value={safeId}
        onChange={setSafeId}
      />
      <ContainerSelect
        containers={bankContainers}
        label={t('safe.forms.bankContainer')}
        value={bankContainerId}
        onChange={setBankContainerId}
      />
      <AmountField value={amountRub} onChange={setAmountRub} />
      <label className="field">
        <span>{t('safe.forms.reason')}</span>
        <input value={reason} onChange={(event) => setReason(event.target.value)} />
      </label>
      <ActorFields
        actorId={actorId}
        approvedById={approvedById}
        onActorIdChange={setActorId}
        onApprovedByIdChange={setApprovedById}
      />
    </FormDialog>
  );
}
