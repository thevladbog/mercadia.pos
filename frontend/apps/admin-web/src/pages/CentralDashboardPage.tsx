import { useGetCentralStatus } from '@mercadia/api-clients-central';
import { Link } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { useAuth } from '@/auth/useAuth.js';
import { canManageCentralUsers } from '@/auth/permissions.js';
import { formatTimestamp } from './reporting-utils.js';

const QUICK_LINKS = [
  {
    label: 'Reporting',
    description: 'Cross-store KPIs and per-store breakdown',
    to: '/central/reporting',
  },
  { label: 'Stores', description: 'Registered store registry', to: '/central/stores' },
  { label: 'Sync', description: 'Synchronized read-model explorer', to: '/central/sync' },
  {
    label: 'Monitoring',
    description: 'Live store-edge terminal monitoring',
    to: '/store/monitoring',
  },
] as const;

export function CentralDashboardPage() {
  const { roles } = useAuth();
  const statusQuery = useGetCentralStatus({
    query: { refetchOnWindowFocus: false },
  });

  const status = statusQuery.data?.status === 200 ? statusQuery.data.data : null;
  const isLoading = statusQuery.isFetching;
  const errorMessage = statusQuery.error != null ? getApiErrorMessage(statusQuery.error) : null;
  const showUsersLink = canManageCentralUsers(roles);

  return (
    <section className="stack reporting-page">
      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>Central Dashboard</h2>
            <p className="muted">Operational status for the central backend region.</p>
          </div>
          <button
            className="secondary"
            disabled={isLoading}
            onClick={() => void statusQuery.refetch()}
            type="button"
          >
            {isLoading ? 'Refreshing…' : 'Refresh'}
          </button>
        </div>
      </div>

      {errorMessage ? (
        <div className="panel error-panel">
          <p className="error">{errorMessage}</p>
        </div>
      ) : null}

      <div className="panel">
        <h3>Central status</h3>
        {statusQuery.isLoading && !status ? (
          <p className="muted">Loading status…</p>
        ) : status ? (
          <dl className="kpi-grid">
            <div>
              <dt>Region</dt>
              <dd>{status.region}</dd>
            </div>
            <div>
              <dt>Status</dt>
              <dd>{status.status}</dd>
            </div>
            <div>
              <dt>Registered stores</dt>
              <dd>{status.storeCount}</dd>
            </div>
            <div>
              <dt>Last updated</dt>
              <dd>{formatTimestamp(status.generatedAt)}</dd>
            </div>
          </dl>
        ) : (
          <p className="muted">No status data.</p>
        )}
      </div>

      <div className="panel">
        <h3>Quick links</h3>
        <div className="stack">
          {QUICK_LINKS.map((link) => (
            <div key={link.to}>
              <Link to={link.to}>{link.label}</Link>
              <p className="muted">{link.description}</p>
            </div>
          ))}
          {showUsersLink ? (
            <div>
              <Link to="/central/users">Users</Link>
              <p className="muted">Central user and role management</p>
            </div>
          ) : null}
        </div>
      </div>
    </section>
  );
}
