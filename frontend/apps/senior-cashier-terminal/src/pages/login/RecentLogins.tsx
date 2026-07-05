import { useTranslation } from 'react-i18next';
import { AvatarChip } from '@mercadia/ui';

import { deriveInitials } from '@/lib/topbar-utils.js';
import { deriveRecentLoginRecency, type RecentLogin } from '@/lib/login-history.js';

export interface RecentLoginsProps {
  logins: RecentLogin[];
  onSelect: (actorId: string) => void;
}

/**
 * "Недавно входили" chip row (design screen 01a). Per plan 020, this is a
 * PER-TERMINAL local history, not a store-wide "who's on shift" feed — see
 * `login-history.ts`'s module doc comment. Tapping a chip only pre-fills the
 * personnel-ID field; it does not log in directly (still requires PIN +
 * credential).
 */
export function RecentLogins({ logins, onSelect }: RecentLoginsProps) {
  const { t } = useTranslation();

  if (logins.length === 0) return null;

  return (
    <div className="sr-login-recent">
      <span className="sr-login-recent-label">{t('auth.wizard.recentlyLoggedIn')}</span>
      <div className="sr-login-recent-list">
        {logins.map((entry) => {
          const recency = deriveRecentLoginRecency(entry.atIso);
          const recencyLabel =
            recency.kind === 'now'
              ? t('auth.wizard.recencyNow')
              : recency.kind === 'time'
                ? recency.hhmm
                : t('auth.wizard.recencyEarlier');
          return (
            <button
              key={`${entry.actorId}-${entry.atIso}`}
              type="button"
              className="sr-login-recent-chip"
              onClick={() => onSelect(entry.actorId)}
            >
              <AvatarChip initials={deriveInitials(entry.actorId)} size="sm" />
              <span className="sr-login-recent-chip-name">{entry.actorId}</span>
              <span className="sr-login-recent-chip-meta">{recencyLabel}</span>
            </button>
          );
        })}
      </div>
    </div>
  );
}
