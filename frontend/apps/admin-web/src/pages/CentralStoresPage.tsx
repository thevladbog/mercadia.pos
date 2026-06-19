import { useListStores } from '@mercadia/api-clients-central';
import { Link } from 'react-router-dom';

import { getApiErrorMessage } from '../auth/api-errors.js';
import { useAuth } from '../auth/useAuth.js';
import { canManageCentralUsers } from '../auth/permissions.js';
import { formatTimestamp } from './reporting-utils.js';

export function CentralStoresPage() {
  const { roles } = useAuth();
  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : null;
  const errorMessage = storesQuery.error != null ? getApiErrorMessage(storesQuery.error) : null;
  const canRegister = canManageCentralUsers(roles);

  return (
    <section className="stack users-page">
      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>Store Registry</h2>
            <p className="muted">Registered stores in the central backend.</p>
          </div>
          <div className="header-actions-inline">
            <button
              className="secondary"
              disabled={storesQuery.isFetching}
              onClick={() => void storesQuery.refetch()}
              type="button"
            >
              {storesQuery.isFetching ? 'Refreshing…' : 'Refresh'}
            </button>
            {canRegister ? (
              <Link className="button-link" to="/central/stores/new">
                Register store
              </Link>
            ) : null}
          </div>
        </div>

        {errorMessage ? (
          <p className="error">{errorMessage}</p>
        ) : storesQuery.isLoading && !stores ? (
          <p className="muted">Loading stores…</p>
        ) : stores && stores.length > 0 ? (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Store ID</th>
                  <th>Name</th>
                  <th>Region</th>
                  <th>Registered</th>
                  <th>Updated</th>
                </tr>
              </thead>
              <tbody>
                {stores.map((store) => (
                  <tr key={store.id}>
                    <td>{store.id}</td>
                    <td>{store.name}</td>
                    <td>{store.region || '—'}</td>
                    <td>{formatTimestamp(store.registeredAt)}</td>
                    <td>{formatTimestamp(store.updatedAt)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="muted">No stores registered yet.</p>
        )}
      </div>
    </section>
  );
}
