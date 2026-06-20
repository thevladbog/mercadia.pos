import { Button } from '@mercadia/ui';
import { Outlet, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { useAuth } from '@/auth/useAuth.js';
import { LanguageSwitcher } from '@/components/LanguageSwitcher.js';
import { AppSidebar } from '@/layout/AppSidebar.js';
import { ThemeToggle } from '@/layout/ThemeToggle.js';

export function AppLayout() {
  const { t } = useTranslation();
  const { logout, roles, userId } = useAuth();
  const location = useLocation();
  const notice = (location.state as { notice?: string } | null)?.notice;

  return (
    <div className="app-shell">
      {notice ? <div className="notice-banner">{notice}</div> : null}
      <div className="app-shell-body">
        <AppSidebar />
        <div className="app-content">
          <header className="app-header">
            <div className="header-actions">
              <ThemeToggle />
              <LanguageSwitcher />
              <div className="user-meta">
                <span>{userId}</span>
                {roles.length > 0 ? <span className="muted">{roles.join(', ')}</span> : null}
              </div>
              <Button variant="secondary" onClick={logout} type="button">
                {t('nav.logout')}
              </Button>
            </div>
          </header>
          <main className="app-main">
            <Outlet />
          </main>
        </div>
      </div>
    </div>
  );
}
