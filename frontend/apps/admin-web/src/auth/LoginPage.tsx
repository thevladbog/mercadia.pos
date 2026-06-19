import { useCreateCentralAuthSession } from '@mercadia/api-clients-central';
import { useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { Navigate, useLocation } from 'react-router-dom';

import { getApiErrorMessage, isUnauthorizedError } from './api-errors.js';
import { useAuth } from './useAuth.js';

export function LoginPage() {
  const { t } = useTranslation();
  const { isAuthenticated, login } = useAuth();
  const location = useLocation();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const mutation = useCreateCentralAuthSession({
    mutation: {
      onSuccess: (response) => {
        if (response.status !== 201) {
          setErrorMessage(t('auth.loginFailed'));
          return;
        }

        const { session } = response.data;
        login(session.userId, session.roles, session.token);
      },
      onError: (error) => {
        if (isUnauthorizedError(error)) {
          setErrorMessage(t('auth.invalidCredentials'));
          return;
        }
        setErrorMessage(getApiErrorMessage(error));
      },
    },
  });

  if (isAuthenticated) {
    const redirectTo = (location.state as { from?: string } | null)?.from ?? '/central/dashboard';
    return <Navigate to={redirectTo} replace />;
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage(null);
    mutation.mutate({ data: { email, password } });
  }

  return (
    <section className="panel login-panel">
      <h1>{t('auth.loginTitle')}</h1>
      <p className="muted">{t('auth.loginSubtitle')}</p>
      <form className="stack" onSubmit={handleSubmit}>
        <label className="field">
          <span>{t('auth.email')}</span>
          <input
            autoComplete="username"
            name="email"
            required
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
          />
        </label>
        <label className="field">
          <span>{t('auth.password')}</span>
          <input
            autoComplete="current-password"
            name="password"
            required
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
          />
        </label>
        {errorMessage ? <p className="error">{errorMessage}</p> : null}
        <button disabled={mutation.isPending} type="submit">
          {mutation.isPending ? t('auth.signingIn') : t('auth.signIn')}
        </button>
      </form>
    </section>
  );
}
