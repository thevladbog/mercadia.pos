import { Link, Outlet, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { useAuth } from '@/auth/useAuth.js';
import { canManageCentralUsers } from '@/auth/permissions.js';
import { LanguageSwitcher } from '@/components/LanguageSwitcher.js';

export function AppLayout() {
  const { t } = useTranslation();
  const { logout, roles, userId } = useAuth();
  const location = useLocation();
  const notice = (location.state as { notice?: string } | null)?.notice;

  return (
    <div className="app-shell">
      {notice ? <div className="notice-banner">{notice}</div> : null}
      <header className="app-header">
        <div>
          <p className="eyebrow">{t('app.eyebrow')}</p>
          <h1>{t('app.title')}</h1>
        </div>
        <div className="header-actions">
          <nav>
            <Link to="/central/dashboard">{t('nav.dashboard')}</Link>
            <Link to="/central/reporting">{t('nav.reporting')}</Link>
            <Link to="/central/stores">{t('nav.stores')}</Link>
            <Link to="/central/sync">{t('nav.sync')}</Link>
            <Link to="/central/catalog">{t('nav.catalog')}</Link>
            <Link to="/store/monitoring">{t('nav.monitoring')}</Link>
            <Link to="/store/safe">{t('nav.safe')}</Link>
            <Link to="/store/eod">{t('nav.eod')}</Link>
            {canManageCentralUsers(roles) ? (
              <Link to="/central/users">{t('nav.users')}</Link>
            ) : null}
          </nav>
          <LanguageSwitcher />
          <div className="user-meta">
            <span>{userId}</span>
            {roles.length > 0 ? <span className="muted">{roles.join(', ')}</span> : null}
          </div>
          <button className="secondary" onClick={logout} type="button">
            {t('nav.logout')}
          </button>
        </div>
      </header>
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  );
}
