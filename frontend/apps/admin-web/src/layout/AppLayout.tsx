import { Link, Outlet, useLocation } from 'react-router-dom';

import { useAuth } from '../auth/AuthProvider.js';
import { canManageCentralUsers } from '../auth/permissions.js';

export function AppLayout() {
  const { logout, roles, userId } = useAuth();
  const location = useLocation();
  const notice = (location.state as { notice?: string } | null)?.notice;

  return (
    <div className="app-shell">
      {notice ? <div className="notice-banner">{notice}</div> : null}
      <header className="app-header">
        <div>
          <p className="eyebrow">Mercadia Admin</p>
          <h1>Central Console</h1>
        </div>
        <div className="header-actions">
          <nav>
            <Link to="/central/reporting">Reporting</Link>
            <Link to="/store/monitoring">Monitoring</Link>
            {canManageCentralUsers(roles) ? <Link to="/central/users">Users</Link> : null}
          </nav>
          <div className="user-meta">
            <span>{userId}</span>
            {roles.length > 0 ? <span className="muted">{roles.join(', ')}</span> : null}
          </div>
          <button className="secondary" onClick={logout} type="button">
            Log out
          </button>
        </div>
      </header>
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  );
}
