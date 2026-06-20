import { useListCentralUsers } from '@mercadia/api-clients-central';
import { Button } from '@mercadia/ui';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { formatTimestamp } from './reporting-utils.js';

export function CentralUsersPage() {
  const { t } = useTranslation();
  const usersQuery = useListCentralUsers();
  const users = usersQuery.data?.status === 200 ? usersQuery.data.data.users : null;
  const errorMessage = usersQuery.error != null ? getApiErrorMessage(usersQuery.error) : null;

  return (
    <section className="stack users-page">
      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>{t('users.title')}</h2>
            <p className="muted">{t('users.subtitle')}</p>
          </div>
          <div className="header-actions-inline">
            <Button
              variant="secondary"
              disabled={usersQuery.isFetching}
              onClick={() => void usersQuery.refetch()}
              type="button"
            >
              {usersQuery.isFetching ? t('common.refreshing') : t('common.refresh')}
            </Button>
            <Link className="button-link" to="/central/users/new">
              {t('users.createUser')}
            </Link>
          </div>
        </div>

        {errorMessage ? (
          <p className="error">{errorMessage}</p>
        ) : usersQuery.isLoading && !users ? (
          <p className="muted">{t('users.loadingUsers')}</p>
        ) : users && users.length > 0 ? (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>{t('users.userId')}</th>
                  <th>{t('users.email')}</th>
                  <th>{t('users.displayName')}</th>
                  <th>{t('users.roles')}</th>
                  <th>{t('users.active')}</th>
                  <th>{t('users.created')}</th>
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
                    <td>{user.active ? t('common.yes') : t('common.no')}</td>
                    <td>{formatTimestamp(user.createdAt)}</td>
                    <td>
                      <Link to={`/central/users/${user.id}`}>{t('common.edit')}</Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="muted">{t('users.noUsers')}</p>
        )}
      </div>
    </section>
  );
}
