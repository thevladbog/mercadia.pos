import { openOperationalDay } from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { useAuth } from '@/auth/useAuth.js';
import { FormDialog } from '@mercadia/ui';
import { createIdempotencyHeaders } from '@/pages/cash-mutation-utils.js';
import { invalidateEodAfterOpen, todayBusinessDate } from '@/pages/eod-mutation-utils.js';
import { storePageHref } from '@/pages/store-routes.js';

const BUSINESS_DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;

type OpenOperationalDayModalProps = {
  storeId: string;
  onClose: () => void;
};

export function OpenOperationalDayModal({ storeId, onClose }: OpenOperationalDayModalProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { userId } = useAuth();
  const openedById = userId ?? '';

  const [businessDate, setBusinessDate] = useState(() => todayBusinessDate());
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: async () =>
      openOperationalDay(
        {
          storeId,
          businessDate: businessDate.trim(),
          openedById,
        },
        { headers: createIdempotencyHeaders() },
      ),
    onSuccess: async (response) => {
      if (response.status === 202) {
        await invalidateEodAfterOpen(queryClient, storeId, response.data.operationalDay.id);
        void navigate(storePageHref('/store/eod', storeId), {
          state: { notice: t('eod.notices.openSuccess') },
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

    const normalizedDate = businessDate.trim();
    if (normalizedDate.length === 0) {
      setErrorMessage(t('eod.forms.validation.businessDateRequired'));
      return;
    }
    if (!BUSINESS_DATE_PATTERN.test(normalizedDate)) {
      setErrorMessage(t('eod.forms.validation.businessDateFormat'));
      return;
    }
    if (openedById.trim().length === 0) {
      setErrorMessage(t('eod.forms.validation.openedByRequired'));
      return;
    }

    mutation.mutate();
  }

  return (
    <FormDialog
      cancelLabel={t('common.cancel')}
      errorMessage={errorMessage}
      isSubmitting={mutation.isPending}
      submitLabel={mutation.isPending ? t('common.submitting') : t('eod.actions.openDay')}
      title={t('eod.forms.openConfirmTitle')}
      onClose={onClose}
      onSubmit={handleSubmit}
    >
      <p className="muted">{t('eod.forms.openConfirmBody')}</p>

      <label className="field">
        <span>{t('eod.businessDate')}</span>
        <input
          required
          type="date"
          value={businessDate}
          onChange={(event) => setBusinessDate(event.target.value)}
        />
      </label>

      <label className="field">
        <span>{t('eod.forms.openedById')}</span>
        <input readOnly value={openedById} />
      </label>
    </FormDialog>
  );
}
