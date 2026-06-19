import { Link, Outlet } from 'react-router-dom';

import { useAuth } from '../auth/AuthProvider.js';

export function AppLayout() {
  const { logout, roles, userId } = useAuth();

  return (
    <div className="app-shell">
      <header className="app-header">
        <div>
          <p className="eyebrow">Mercadia Admin</p>
          <h1>Central Console</h1>
        </div>
        <div className="header-actions">
          <nav>
            <Link to="/central/reporting">Reporting</Link>
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
