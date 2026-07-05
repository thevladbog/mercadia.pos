import { useTranslation } from 'react-i18next';
import { Button } from '@mercadia/ui';

import type { StaffCredentialKind, StaffCredentialRead } from '@/auth/ibutton.js';

import { IButtonGraphic } from './icons.js';
import { IdentityMiniCard } from './IdentityMiniCard.js';

const CREDENTIAL_KINDS: StaffCredentialKind[] = ['ibutton', 'msr_card', 'barcode_card'];

export interface CredentialStepProps {
  personnelId: string;
  credentialKind: StaffCredentialKind;
  onCredentialKindChange: (kind: StaffCredentialKind) => void;
  credentialStatus: 'idle' | 'waiting' | 'detected' | 'error';
  credentialRead: StaffCredentialRead | null;
  onReadCredential: () => void;
  isSubmitting: boolean;
  isBlocked: boolean;
  attempts: number;
  maxAttempts: number;
  authError: string;
  credentialError: string;
  onCancel: () => void;
}

/**
 * Step 3 of the login wizard (design screen 01c): credential-kind selector
 * + read trigger. The design shows no explicit "read" button — reading
 * appears to auto-advance. This implementation keeps an EXPLICIT
 * "read credential" button instead (same as the pre-wizard `LoginPage` and
 * `CredentialEnrollmentPage.tsx`), because auto-triggering a hardware read
 * the instant this step mounts is surprising: it would fire a device
 * command before the operator has necessarily picked the right credential
 * kind or is ready to present their key/card, with no easy way to cancel a
 * read already in flight. Once a read succeeds, `LoginPage.tsx` calls the
 * real `login()` immediately and automatically — there is still no visible
 * "submit" button, matching the design.
 */
export function CredentialStep({
  personnelId,
  credentialKind,
  onCredentialKindChange,
  credentialStatus,
  credentialRead,
  onReadCredential,
  isSubmitting,
  isBlocked,
  attempts,
  maxAttempts,
  authError,
  credentialError,
  onCancel,
}: CredentialStepProps) {
  const { t } = useTranslation();
  const busy = isSubmitting || credentialStatus === 'waiting';

  return (
    <div className="sr-login-right-content">
      <IdentityMiniCard
        variant="badges"
        personnelId={personnelId}
        credentialKind={credentialKind}
        credentialConfirmed={credentialStatus === 'detected'}
      />

      <div className="sr-login-step-heading">
        <span className="sr-login-step-heading-label">
          {t('auth.wizard.stepOfTotal', { current: 3, total: 3 })} ·{' '}
          {t('auth.wizard.stepCredentialTitle')}
        </span>
        <h2 className="sr-login-step-heading-title">{t('auth.wizard.credentialPrompt')}</h2>
      </div>

      <div className="sr-login-credential-graphic">
        <IButtonGraphic />
        <span className="sr-login-credential-graphic-label">
          {t(`auth.credentialKinds.${credentialKind}`)}
        </span>
      </div>

      <div
        className="sr-credential-options"
        role="radiogroup"
        aria-label={t('auth.credentialKind')}
        style={{ width: '100%' }}
      >
        {CREDENTIAL_KINDS.map((kind) => (
          <Button
            key={kind}
            type="button"
            variant={credentialKind === kind ? 'primary' : 'secondary'}
            onClick={() => onCredentialKindChange(kind)}
            disabled={busy || isBlocked}
            aria-pressed={credentialKind === kind}
          >
            {t(`auth.credentialKinds.${kind}`)}
          </Button>
        ))}
      </div>

      <Button type="button" onClick={onReadCredential} disabled={busy || isBlocked}>
        {credentialStatus === 'waiting'
          ? t('auth.credentialWaiting')
          : credentialRead
            ? t('auth.rereadCredential')
            : t('auth.readCredential')}
      </Button>

      {isBlocked && <p className="sr-field-error">{t('auth.blocked')}</p>}
      {!isBlocked && authError && <p className="sr-field-error">{authError}</p>}
      {credentialError && <p className="sr-field-error">{credentialError}</p>}
      {attempts > 0 && !isBlocked && (
        <p className="sr-field-error">
          {t('auth.attemptsRemaining', { count: maxAttempts - attempts })}
        </p>
      )}

      <div className="sr-login-alt-method">
        <div className="sr-login-alt-method-label">{t('auth.wizard.alternativeMethod')}</div>
        <div>{t('auth.wizard.alternativeMethodHint')}</div>
      </div>

      <Button type="button" variant="secondary" onClick={onCancel} disabled={busy}>
        {t('auth.wizard.cancelLogin')}
      </Button>
    </div>
  );
}
