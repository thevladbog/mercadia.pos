import {
  ApiError,
  getListCentralUsersQueryKey,
  useCreateCentralUser,
  type CreateCentralUserBody,
} from '@mercadia/api-clients-central';
import { Button } from '@mercadia/ui';
import { useQueryClient } from '@tanstack/react-query';
import { useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { CENTRAL_ROLE_VIEWER } from '@/auth/permissions.js';
import { CentralRoleFields, PageBackLink } from './users-shared.js';

export function CreateCentralUserPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [userId, setUserId] = useState('');
  const [email, setEmail] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [password, setPassword] = useState('');
  const [roles, setRoles] = useState<string[]>([CENTRAL_ROLE_VIEWER]);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const mutation = useCreateCentralUser({
    mutation: {
      onSuccess: async (response) => {
        if (response.status !== 201) {
          setErrorMessage(t('common.unexpectedError'));
          return;
        }
        await queryClient.invalidateQueries({ queryKey: getListCentralUsersQueryKey() });
        void navigate('/central/users');
      },
      onError: (error) => {
        if (error instanceof ApiError && error.status === 409) {
          setErrorMessage(getApiErrorMessage(error));
          return;
        }
        setErrorMessage(getApiErrorMessage(error));
      },
    },
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage(null);

    if (roles.length === 0) {
      setErrorMessage(t('users.roles'));
      return;
    }

    const payload: CreateCentralUserBody = {
      userId: userId.trim(),
      email: email.trim(),
      password,
      roles,
      ...(displayName.trim() ? { displayName: displayName.trim() } : {}),
    };

    mutation.mutate({ data: payload });
  }

  return (
    <section className="stack users-page">
      <PageBackLink />
      <div className="panel login-panel">
        <h2>{t('users.createTitle')}</h2>
        <p className="muted">{t('users.subtitle')}</p>
        <form className="stack" onSubmit={handleSubmit}>
          <label className="field">
            <span>{t('users.userId')}</span>
            <input required value={userId} onChange={(event) => setUserId(event.target.value)} />
          </label>
          <label className="field">
            <span>{t('users.email')}</span>
            <input
              required
              type="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
            />
          </label>
          <label className="field">
            <span>{t('users.displayName')}</span>
            <input value={displayName} onChange={(event) => setDisplayName(event.target.value)} />
          </label>
          <label className="field">
            <span>{t('users.password')}</span>
            <input
              required
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />
          </label>
          <CentralRoleFields roles={roles} onChange={setRoles} />
          {errorMessage ? <p className="error">{errorMessage}</p> : null}
          <div className="form-actions">
            <Button disabled={mutation.isPending} type="submit">
              {mutation.isPending ? t('users.creating') : t('users.createUser')}
            </Button>
          </div>
        </form>
      </div>
    </section>
  );
}
