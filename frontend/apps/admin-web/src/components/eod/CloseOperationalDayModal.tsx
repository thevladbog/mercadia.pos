import type { GetOperationalDaySummary200 } from '@mercadia/api-clients-store-edge';
import { closeOperationalDay, useGetOperationalDaySummary } from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useEffect, useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { useAuth } from '@/auth/useAuth.js';
import { CashModal } from '@/components/cash/CashModal.js';
import { actorsMustDiffer, createIdempotencyHeaders } from '@/pages/cash-mutation-utils.js';
import {
  analyzeCloseReadiness,
  formatBlockerMessage,
  formatBlockerSeverity,
  invalidateEodQueries,
} from '@/pages/eod-mutation-utils.js';
import { storePageHref } from '@/pages/store-routes.js';

type CloseOperationalDayModalProps = {
  storeId: string;
  operationalDayId: string;
  onClose: () => void;
};

export function CloseOperationalDayModal({
  storeId,
  operationalDayId,
  onClose,
}: CloseOperationalDayModalProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { userId } = useAuth();
  const closedById = userId ?? '';

  const [overrideNoSales, setOverrideNoSales] = useState(false);
  const [overrideActorId, setOverrideActorId] = useState('');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const summaryQuery = useGetOperationalDaySummary(operationalDayId, {
    query: { enabled: operationalDayId.length > 0 },
  });
  const { refetch, isFetching } = summaryQuery;
  const summary: GetOperationalDaySummary200 | null =
    summaryQuery.data?.status === 200 ? summaryQuery.data.data : null;
  const blockers = summary?.blockers ?? [];
  const readiness = analyzeCloseReadiness(blockers);
  const requiresOverride = readiness.canCloseWithOverride;

  useEffect(() => {
    void refetch();
  }, [operationalDayId, refetch]);

  const mutation = useMutation({
    mutationFn: async () =>
      closeOperationalDay(
        operationalDayId,
        {
          closedById,
          overrideNoSales: requiresOverride ? overrideNoSales : undefined,
          overrideActorId: requiresOverride && overrideNoSales ? overrideActorId.trim() : undefined,
        },
        { headers: createIdempotencyHeaders() },
      ),
    onSuccess: async (response) => {
      if (response.status === 202) {
        await invalidateEodQueries(queryClient, storeId, operationalDayId);
        void navigate(storePageHref('/store/eod', storeId), {
          state: { notice: t('eod.notices.closeSuccess') },
        });
        onClose();
        return;
      }
      setErrorMessage(t('common.unexpectedError'));
    },
    onError: (error) => {
      setErrorMessage(getApiErrorMessage(error));
      void refetch();
    },
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage(null);

    if (closedById.trim().length === 0) {
      setErrorMessage(t('eod.forms.validation.closedByRequired'));
      return;
    }

    if (requiresOverride) {
      if (!overrideNoSales) {
        setErrorMessage(t('eod.forms.validation.overrideRequired'));
        return;
      }
      if (overrideActorId.trim().length === 0) {
        setErrorMessage(t('eod.forms.validation.overrideActorRequired'));
        return;
      }
      if (!actorsMustDiffer(closedById, overrideActorId)) {
        setErrorMessage(t('eod.forms.validation.selfOverride'));
        return;
      }
    }

    mutation.mutate();
  }

  return (
    <CashModal
      errorMessage={errorMessage}
      isSubmitting={mutation.isPending || isFetching}
      submitLabel={t('eod.actions.closeDay')}
      title={t('eod.forms.confirmTitle')}
      onClose={onClose}
      onSubmit={handleSubmit}
    >
      <p className="muted">{t('eod.forms.confirmBody')}</p>

      {isFetching && !summary ? (
        <p className="muted">{t('common.loading')}</p>
      ) : blockers.length > 0 ? (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>{t('eod.severity')}</th>
                <th>{t('eod.code')}</th>
                <th>{t('eod.message')}</th>
                <th>{t('eod.reference')}</th>
              </tr>
            </thead>
            <tbody>
              {blockers.map((blocker) => (
                <tr key={`${blocker.code}-${blocker.referenceId ?? blocker.message}`}>
                  <td>{formatBlockerSeverity(blocker.severity, t)}</td>
                  <td>{blocker.code}</td>
                  <td>{formatBlockerMessage(blocker, t)}</td>
                  <td>{blocker.referenceId ?? t('common.emDash')}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <p className="muted">{t('eod.noBlockers')}</p>
      )}

      <label className="field">
        <span>{t('eod.forms.closedById')}</span>
        <input readOnly value={closedById} />
      </label>

      {requiresOverride ? (
        <div className="stack">
          <label className="field checkbox-field">
            <input
              checked={overrideNoSales}
              type="checkbox"
              onChange={(event) => setOverrideNoSales(event.target.checked)}
            />
            <span>{t('eod.forms.overrideNoSales')}</span>
          </label>
          <label className="field">
            <span>{t('eod.forms.overrideActorId')}</span>
            <input
              required={overrideNoSales}
              value={overrideActorId}
              onChange={(event) => setOverrideActorId(event.target.value)}
            />
          </label>
          <p className="muted form-hint">{t('safe.forms.actorHint')}</p>
        </div>
      ) : null}
    </CashModal>
  );
}
