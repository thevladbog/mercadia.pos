import { useTranslation } from 'react-i18next';
import { Button, Field, Input, Label } from '@mercadia/ui';

import type { RecentLogin } from '@/lib/login-history.js';

import { LockIcon } from './icons.js';
import { RecentLogins } from './RecentLogins.js';

export interface PersonnelIdStepProps {
  personnelId: string;
  onPersonnelIdChange: (value: string) => void;
  onContinue: () => void;
  recentLogins: RecentLogin[];
  onSelectRecent: (actorId: string) => void;
}

/** Step 1 of the login wizard (design screen 01a): personnel-ID entry + recent-login chips. */
export function PersonnelIdStep({
  personnelId,
  onPersonnelIdChange,
  onContinue,
  recentLogins,
  onSelectRecent,
}: PersonnelIdStepProps) {
  const { t } = useTranslation();
  const canContinue = personnelId.trim().length > 0;

  return (
    <div className="sr-login-right-content">
      <div className="sr-login-idle-icon">
        <LockIcon width={40} height={40} />
      </div>
      <h2 className="sr-login-idle-heading">{t('auth.wizard.idlePrompt')}</h2>
      <p className="sr-login-idle-description">{t('auth.wizard.idleDescription')}</p>

      <form
        className="sr-login-personnel-form"
        onSubmit={(event) => {
          event.preventDefault();
          if (canContinue) onContinue();
        }}
      >
        <Field>
          <Label>{t('auth.personnelId')}</Label>
          <Input
            value={personnelId}
            onChange={(event) => onPersonnelIdChange(event.target.value)}
            placeholder={t('auth.personnelIdPlaceholder')}
            autoFocus
          />
        </Field>
        <Button type="submit" disabled={!canContinue}>
          {t('auth.wizard.continue')}
        </Button>
      </form>

      <RecentLogins logins={recentLogins} onSelect={onSelectRecent} />

      <p className="sr-login-security-note">{t('auth.wizard.attemptsLoggedNote')}</p>
    </div>
  );
}
