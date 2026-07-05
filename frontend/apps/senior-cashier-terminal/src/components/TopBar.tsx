import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { AvatarChip, Badge, Button } from '@mercadia/ui';

import { useAuth } from '@/auth/AuthProvider.js';
import {
  deriveElapsed,
  deriveInitials,
  deriveLoginAt,
  formatCountdown,
  formatLoginTime,
} from '@/lib/topbar-utils.js';
import { LanguageSwitcher } from '@/components/LanguageSwitcher.js';
import { ThemeToggle } from '@/components/ThemeToggle.js';

import './TopBar.css';

export interface TopBarProps {
  /** Navigate to the shift-handover flow. */
  onHandover: () => void;
  /** Lock/log out the current session. Callers pass their own existing handler
   * (may differ from the shared `logout()` — see `ShiftHandoverPage`). */
  onLock: () => void;
  /** Omit to hide the "operations" pill — no data source exists yet for most pages. */
  operationsCount?: number;
  /** Omit to hide the "alerts" pill — no data source exists yet for most pages. */
  alertsCount?: number;
}

/**
 * App-wide top bar (plan 019, Phase 2 of the senior-cashier-terminal
 * redesign). Reads `session`/`remaining` from `useAuth()` directly since
 * every consumer already sits inside `AuthProvider` — no need to pass them
 * down as props.
 */
export function TopBar({ onHandover, onLock, operationsCount, alertsCount }: TopBarProps) {
  const { t } = useTranslation();
  const { session, remaining } = useAuth();
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(id);
  }, []);

  if (!session) return null;

  const initials = deriveInitials(session.actorId);
  const loginAt = deriveLoginAt(session.expiresAt);
  const elapsed = deriveElapsed(loginAt, now);
  const role = session.roles[0] ?? '';

  return (
    <header className="sr-topbar">
      <div className="sr-topbar-brand">
        <span className="sr-topbar-logo" aria-hidden="true">
          M
        </span>
        <span className="sr-topbar-wordmark">{t('app.brand')}</span>
      </div>

      <div className="sr-topbar-identity">
        <AvatarChip initials={initials} size="sm" />
        <div className="sr-topbar-identity-text">
          <span className="sr-topbar-identity-label">{t('topbar.authorized')}</span>
          <span className="sr-topbar-identity-line">
            {session.actorId} · {role} · {t('topbar.since', { time: formatLoginTime(loginAt) })} (
            {elapsed.hours}
            {t('dashboard.hours')} {elapsed.minutes}
            {t('dashboard.minutes')})
          </span>
        </div>
      </div>

      <div className="sr-topbar-actions">
        <div className="sr-topbar-pills">
          {operationsCount !== undefined && (
            <Badge variant="outline">{t('topbar.operations', { count: operationsCount })}</Badge>
          )}
          {alertsCount !== undefined && (
            <Badge variant={alertsCount > 0 ? 'danger' : 'outline'}>
              {t('topbar.alerts', { count: alertsCount })}
            </Badge>
          )}
          <Badge variant="outline">
            {t('topbar.autoLock', { value: formatCountdown(remaining) })}
          </Badge>
        </div>

        <Button variant="secondary" size="sm" onClick={onHandover}>
          {t('topbar.handover')}
        </Button>
        <Button variant="primary" size="sm" onClick={onLock}>
          {t('topbar.lock')}
        </Button>

        <div className="sr-topbar-utility">
          <LanguageSwitcher />
          <ThemeToggle />
        </div>
      </div>
    </header>
  );
}
