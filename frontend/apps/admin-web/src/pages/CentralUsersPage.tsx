import { useListCentralUsers } from '@mercadia/api-clients-central';
import { Link } from 'react-router-dom';

import { getApiErrorMessage } from '../auth/api-errors.js';
import { formatTimestamp } from './reporting-utils.js';

export function CentralUsersPage() {
  const usersQuery = useListCentralUsers();
  const users = usersQuery.data?.status === 200 ? usersQuery.data.data.users : null;
  const errorMessage = usersQuery.error != null ? getApiErrorMessage(usersQuery.error) : null;

  return (
    <section className="stack users-page">
      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>Central Users</h2>
            <p className="muted">Manage central admin accounts and roles.</p>
          </div>
          <div className="header-actions-inline">
            <button
              className="secondary"
              disabled={usersQuery.isFetching}
              onClick={() => void usersQuery.refetch()}
              type="button"
            >
              {usersQuery.isFetching ? 'Refreshing…' : 'Refresh'}
            </button>
            <Link className="button-link" to="/central/users/new">
              Create user
            </Link>
          </div>
        </div>

        {errorMessage ? (
          <p className="error">{errorMessage}</p>
        ) : usersQuery.isLoading && !users ? (
          <p className="muted">Loading users…</p>
        ) : users && users.length > 0 ? (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>User ID</th>
                  <th>Email</th>
                  <th>Display name</th>
                  <th>Roles</th>
                  <th>Active</th>
                  <th>Created</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {users.map((user) => (
                  <tr key={user.id}>
                    <td>{user.id}</td>
                    <td>{user.email}</td>
                    <td>{user.displayName}</td>
                    <td>{user.roles.join(', ')}</td>
                    <td>{user.active ? 'Yes' : 'No'}</td>
                    <td>{formatTimestamp(user.createdAt)}</td>
                    <td>
                      <Link to={`/central/users/${user.id}`}>Edit</Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="muted">No central users found.</p>
        )}
      </div>
    </section>
  );
}
