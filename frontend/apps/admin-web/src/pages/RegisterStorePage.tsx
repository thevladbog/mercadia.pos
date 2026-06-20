import {
  ApiError,
  getListStoresQueryKey,
  registerStore,
  type RegisterStoreBody,
} from '@mercadia/api-clients-central';
import { Button } from '@mercadia/ui';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useNavigate } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';

export function RegisterStorePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [storeId, setStoreId] = useState('');
  const [name, setName] = useState('');
  const [region, setRegion] = useState('');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: (payload: RegisterStoreBody) =>
      registerStore(payload, {
        headers: { 'Idempotency-Key': crypto.randomUUID() },
      }),
    onSuccess: async (response) => {
      if (response.status === 202) {
        await queryClient.invalidateQueries({ queryKey: getListStoresQueryKey() });
        void navigate('/central/stores');
        return;
      }
      if (response.status === 409) {
        setErrorMessage(t('common.unexpectedError'));
        return;
      }
      if (response.status === 403) {
        setErrorMessage(t('auth.adminRequiredNotice'));
        return;
      }
      setErrorMessage(t('common.unexpectedError'));
    },
    onError: (error) => {
      if (error instanceof ApiError && error.status === 409) {
        setErrorMessage(getApiErrorMessage(error));
        return;
      }
      if (error instanceof ApiError && error.status === 403) {
        setErrorMessage(t('auth.adminRequiredNotice'));
        return;
      }
      setErrorMessage(getApiErrorMessage(error));
    },
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage(null);

    const payload: RegisterStoreBody = {
      storeId: storeId.trim(),
      name: name.trim(),
      ...(region.trim() ? { region: region.trim() } : {}),
    };

    mutation.mutate(payload);
  }

  return (
    <section className="stack users-page">
      <p className="page-back">
        <Link to="/central/stores">{t('nav.stores')}</Link>
      </p>
      <div className="panel login-panel">
        <h2>{t('stores.registerTitle')}</h2>
        <p className="muted">{t('stores.registerSubtitle')}</p>
        <form className="stack" onSubmit={handleSubmit}>
          <label className="field">
            <span>{t('stores.storeIdField')}</span>
            <input required value={storeId} onChange={(event) => setStoreId(event.target.value)} />
          </label>
          <label className="field">
            <span>{t('stores.storeName')}</span>
            <input required value={name} onChange={(event) => setName(event.target.value)} />
          </label>
          <label className="field">
            <span>{t('stores.regionField')}</span>
            <input value={region} onChange={(event) => setRegion(event.target.value)} />
          </label>
          {errorMessage ? <p className="error">{errorMessage}</p> : null}
          <div className="form-actions">
            <Button disabled={mutation.isPending} type="submit">
              {mutation.isPending ? t('stores.registering') : t('stores.submitRegister')}
            </Button>
          </div>
        </form>
      </div>
    </section>
  );
}
