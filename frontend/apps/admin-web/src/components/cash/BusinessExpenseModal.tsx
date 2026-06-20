import type { ListCashBalances200BalancesItem } from '@mercadia/api-clients-store-edge';
import { createBusinessExpense } from '@mercadia/api-clients-store-edge';
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

type BusinessExpenseModalProps = {
  storeId: string;
  balances: ListCashBalances200BalancesItem[];
  onClose: () => void;
};

export function BusinessExpenseModal({ storeId, balances, onClose }: BusinessExpenseModalProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const safeContainers = useMemo(() => containersByType(balances, 'safe'), [balances]);
  const defaultSafe = firstContainerByType(balances, 'safe');

  const [safeId, setSafeId] = useState(defaultSafe?.containerId ?? '');
  const [payeeId, setPayeeId] = useState('');
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
      return createBusinessExpense(
        storeId,
        {
          safeId,
          payeeId: payeeId.trim(),
          reason: reason.trim(),
          amountMinor,
          actorId: actorId.trim(),
          approvedById: approvedById.trim(),
        },
        { headers: createIdempotencyHeaders() },
      );
    },
    onSuccess: async (response) => {
      if (response.status === 202) {
        await invalidateSafeQueries(queryClient, storeId);
        void navigate(storePageHref('/store/safe', storeId), {
          state: { notice: t('safe.notices.expenseSuccess') },
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
    if (!safeId || payeeId.trim().length === 0) {
      setErrorMessage(t('safe.forms.validation.payee'));
      return;
    }
    if (reason.trim().length === 0) {
      setErrorMessage(t('safe.forms.validation.reasonRequired'));
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
      title={t('safe.actions.businessExpense')}
      onClose={onClose}
      onSubmit={handleSubmit}
    >
      <ContainerSelect
        containers={safeContainers}
        label={t('safe.forms.safeContainer')}
        value={safeId}
        onChange={setSafeId}
      />
      <label className="field">
        <span>{t('safe.forms.payeeId')}</span>
        <input required value={payeeId} onChange={(event) => setPayeeId(event.target.value)} />
      </label>
      <AmountField value={amountRub} onChange={setAmountRub} />
      <label className="field">
        <span>{t('safe.forms.reason')}</span>
        <input required value={reason} onChange={(event) => setReason(event.target.value)} />
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
