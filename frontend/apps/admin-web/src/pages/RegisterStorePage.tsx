import {
  ApiError,
  getListStoresQueryKey,
  registerStore,
  type RegisterStoreBody,
} from '@mercadia/api-clients-central';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';

export function RegisterStorePage() {
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
        setErrorMessage('Store already registered or idempotency conflict');
        return;
      }
      if (response.status === 403) {
        setErrorMessage('Central admin role is required to register stores');
        return;
      }
      setErrorMessage('Store registration failed');
    },
    onError: (error) => {
      if (error instanceof ApiError && error.status === 409) {
        setErrorMessage('Store already registered or idempotency conflict');
        return;
      }
      if (error instanceof ApiError && error.status === 403) {
        setErrorMessage('Central admin role is required to register stores');
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
        <Link to="/central/stores">Back to stores</Link>
      </p>
      <div className="panel login-panel">
        <h2>Register Store</h2>
        <p className="muted">Add a store to the central registry for sync and monitoring.</p>
        <form className="stack" onSubmit={handleSubmit}>
          <label className="field">
            <span>Store ID</span>
            <input required value={storeId} onChange={(event) => setStoreId(event.target.value)} />
          </label>
          <label className="field">
            <span>Name</span>
            <input required value={name} onChange={(event) => setName(event.target.value)} />
          </label>
          <label className="field">
            <span>Region (optional)</span>
            <input value={region} onChange={(event) => setRegion(event.target.value)} />
          </label>
          {errorMessage ? <p className="error">{errorMessage}</p> : null}
          <div className="form-actions">
            <button disabled={mutation.isPending} type="submit">
              {mutation.isPending ? 'Registering…' : 'Register store'}
            </button>
          </div>
        </form>
      </div>
    </section>
  );
}
