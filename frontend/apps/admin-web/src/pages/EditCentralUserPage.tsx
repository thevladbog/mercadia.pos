import {
  ApiError,
  getGetCentralUserQueryKey,
  getListCentralUsersQueryKey,
  useGetCentralUser,
  useUpdateCentralUser,
  type GetCentralUser200User,
  type UpdateCentralUserBody,
} from '@mercadia/api-clients-central';
import { useQueryClient } from '@tanstack/react-query';
import { useState, type FormEvent } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';

import { getApiErrorMessage } from '../auth/AuthProvider.js';
import { CentralRoleFields, PageBackLink } from './users-shared.js';

type EditCentralUserFormProps = {
  user: GetCentralUser200User;
  userId: string;
};

function EditCentralUserForm({ user, userId }: EditCentralUserFormProps) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [displayName, setDisplayName] = useState(user.displayName);
  const [password, setPassword] = useState('');
  const [roles, setRoles] = useState<string[]>(user.roles);
  const [active, setActive] = useState(user.active);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const mutation = useUpdateCentralUser({
    mutation: {
      onSuccess: async (response) => {
        if (response.status !== 200) {
          setErrorMessage('User update failed');
          return;
        }
        await Promise.all([
          queryClient.invalidateQueries({ queryKey: getListCentralUsersQueryKey() }),
          queryClient.invalidateQueries({ queryKey: getGetCentralUserQueryKey(userId) }),
        ]);
        void navigate('/central/users');
      },
      onError: (error) => {
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

    const payload: UpdateCentralUserBody = {
      displayName: displayName.trim(),
      roles,
      active,
      ...(password.trim() ? { password: password } : {}),
    };

    mutation.mutate({ userId, data: payload });
  }

  return (
    <div className="panel login-panel">
      <h2>Edit Central User</h2>
      <p className="muted">Update roles, active state, or password for {user.email}.</p>
      <form className="stack" onSubmit={handleSubmit}>
        <label className="field">
          <span>User ID</span>
          <p className="readonly-field">{user.id}</p>
        </label>
        <label className="field">
          <span>Email</span>
          <p className="readonly-field">{user.email}</p>
        </label>
        <label className="field">
          <span>Display name</span>
          <input
            required
            value={displayName}
            onChange={(event) => setDisplayName(event.target.value)}
          />
        </label>
        <label className="field">
          <span>New password (optional)</span>
          <input
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
          />
        </label>
        <CentralRoleFields roles={roles} onChange={setRoles} />
        <label className="checkbox-field">
          <input
            checked={active}
            type="checkbox"
            onChange={(event) => setActive(event.target.checked)}
          />
          <span>Active</span>
        </label>
        {errorMessage ? <p className="error">{errorMessage}</p> : null}
        <div className="form-actions">
          <button disabled={mutation.isPending} type="submit">
            {mutation.isPending ? 'Saving…' : 'Save changes'}
          </button>
        </div>
      </form>
    </div>
  );
}

export function EditCentralUserPage() {
  const { userId = '' } = useParams();
  const userQuery = useGetCentralUser(userId, { query: { enabled: userId.length > 0 } });

  const user = userQuery.data?.status === 200 ? userQuery.data.data.user : null;
  const notFound = userQuery.data?.status === 404;

  if (!userId) {
    return (
      <section className="panel">
        <p className="error">Missing user ID.</p>
        <PageBackLink />
      </section>
    );
  }

  if (userQuery.isLoading && !user && !notFound) {
    return (
      <section className="panel">
        <p className="muted">Loading user…</p>
      </section>
    );
  }

  if (notFound || (userQuery.error instanceof ApiError && userQuery.error.status === 404)) {
    return (
      <section className="panel">
        <h2>User not found</h2>
        <p className="muted">No central user exists with ID {userId}.</p>
        <p>
          <Link to="/central/users">Back to users</Link>
        </p>
      </section>
    );
  }

  if (userQuery.error != null) {
    return (
      <section className="panel">
        <p className="error">{getApiErrorMessage(userQuery.error)}</p>
        <PageBackLink />
      </section>
    );
  }

  if (!user) {
    return (
      <section className="panel">
        <p className="muted">No user data.</p>
        <PageBackLink />
      </section>
    );
  }

  return (
    <section className="stack users-page">
      <PageBackLink />
      <EditCentralUserForm key={user.id} user={user} userId={userId} />
    </section>
  );
}
