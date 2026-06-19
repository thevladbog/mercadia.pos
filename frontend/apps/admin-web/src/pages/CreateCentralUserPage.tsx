import {
  ApiError,
  getListCentralUsersQueryKey,
  useCreateCentralUser,
  type CreateCentralUserBody,
} from '@mercadia/api-clients-central';
import { useQueryClient } from '@tanstack/react-query';
import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { CENTRAL_ROLE_VIEWER } from '@/auth/permissions.js';
import { CentralRoleFields, PageBackLink } from './users-shared.js';

export function CreateCentralUserPage() {
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
          setErrorMessage('User creation failed');
          return;
        }
        await queryClient.invalidateQueries({ queryKey: getListCentralUsersQueryKey() });
        void navigate('/central/users');
      },
      onError: (error) => {
        if (error instanceof ApiError && error.status === 409) {
          setErrorMessage('A user with this email already exists');
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
      setErrorMessage('Select at least one role');
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
        <h2>Create Central User</h2>
        <p className="muted">Add a new central admin or viewer account.</p>
        <form className="stack" onSubmit={handleSubmit}>
          <label className="field">
            <span>User ID</span>
            <input required value={userId} onChange={(event) => setUserId(event.target.value)} />
          </label>
          <label className="field">
            <span>Email</span>
            <input
              required
              type="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
            />
          </label>
          <label className="field">
            <span>Display name (optional)</span>
            <input value={displayName} onChange={(event) => setDisplayName(event.target.value)} />
          </label>
          <label className="field">
            <span>Password</span>
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
            <button disabled={mutation.isPending} type="submit">
              {mutation.isPending ? 'Creating…' : 'Create user'}
            </button>
          </div>
        </form>
      </div>
    </section>
  );
}
