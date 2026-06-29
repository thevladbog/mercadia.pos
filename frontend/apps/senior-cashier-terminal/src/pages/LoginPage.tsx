import { useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button, Input, Field, Label } from '@mercadia/ui';

import { useAuth } from '@/auth/AuthProvider.js';
import {
  readStaffCredential,
  type StaffCredentialKind,
  type StaffCredentialRead,
} from '@/auth/ibutton.js';
import { useIdleTimer } from '@/lib/use-idle-timer.js';

const MAX_ATTEMPTS = 5;
const ATTEMPTS_KEY = 'mercadia.sr-terminal.login-attempts';
const CREDENTIAL_KINDS: StaffCredentialKind[] = ['ibutton', 'msr_card', 'barcode_card'];

function loadAttempts(): number {
  try {
    return parseInt(sessionStorage.getItem(ATTEMPTS_KEY) ?? '0', 10) || 0;
  } catch {
    return 0;
  }
}

function saveAttempts(n: number): void {
  try {
    sessionStorage.setItem(ATTEMPTS_KEY, String(n));
  } catch {
    /* noop */
  }
}

export function LoginPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { login, session } = useAuth();

  const [personnelId, setPersonnelId] = useState('');
  const [pin, setPin] = useState('');
  const [credentialKind, setCredentialKind] = useState<StaffCredentialKind>('ibutton');
  const [credentialRead, setCredentialRead] = useState<StaffCredentialRead | null>(null);
  const [credentialStatus, setCredentialStatus] = useState<
    'idle' | 'waiting' | 'detected' | 'error'
  >('idle');
  const [error, setError] = useState('');
  const [attempts, setAttempts] = useState(loadAttempts);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useIdleTimer();

  const handleCredentialKind = useCallback((kind: StaffCredentialKind) => {
    setCredentialKind(kind);
    setCredentialRead(null);
    setCredentialStatus('idle');
    setError('');
  }, []);

  const handleCredentialRead = useCallback(async () => {
    setCredentialStatus('waiting');
    setError('');
    try {
      const nextCredentialRead = await readStaffCredential(credentialKind);
      setCredentialRead(nextCredentialRead);
      setCredentialStatus('detected');
    } catch {
      setCredentialRead(null);
      setCredentialStatus('error');
      setError(t('auth.credentialError'));
    }
  }, [credentialKind, t]);

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      if (isSubmitting) return;

      if (attempts >= MAX_ATTEMPTS) {
        setError(t('auth.blocked'));
        return;
      }

      if (!personnelId || !pin) {
        setError(t('auth.invalidCredentials'));
        return;
      }

      if (!credentialRead || credentialRead.factor.kind !== credentialKind) {
        setError(t('auth.credentialRequired'));
        return;
      }

      setIsSubmitting(true);
      setError('');

      try {
        const sess = await login(personnelId, pin, credentialRead.factor);
        const target =
          sess.roles.includes('senior_cashier') || sess.roles.includes('admin')
            ? '/dashboard'
            : '/monitoring';
        navigate(target, { replace: true });
      } catch (err) {
        if (err instanceof Error && err.message === 'Invalid credentials') {
          const next = attempts + 1;
          setAttempts(next);
          saveAttempts(next);
        }
        setError(t('auth.invalidCredentials'));
      } finally {
        setIsSubmitting(false);
      }
    },
    [personnelId, pin, credentialKind, credentialRead, login, navigate, t, attempts, isSubmitting],
  );

  if (session) {
    const target =
      session.roles.includes('senior_cashier') || session.roles.includes('admin')
        ? '/dashboard'
        : '/monitoring';
    navigate(target, { replace: true });
    return null;
  }

  return (
    <div className="sr-terminal-shell" style={{ alignItems: 'center', justifyContent: 'center' }}>
      <div
        className="sr-panel"
        style={{ width: 'min(400px, calc(100vw - 2rem))', padding: '2rem' }}
      >
        <div style={{ textAlign: 'center', marginBottom: '1.5rem' }}>
          <h1 style={{ margin: 0, fontSize: '1.5rem', color: 'var(--ui-accent)' }}>
            {t('auth.loginTitle')}
          </h1>
          <p className="muted" style={{ fontSize: '0.85rem', marginTop: '0.25rem' }}>
            MERCADIA · SR. CASHIER
          </p>
        </div>

        <form onSubmit={handleSubmit} className="sr-form">
          <Field>
            <Label>{t('auth.personnelId')}</Label>
            <Input
              value={personnelId}
              onChange={(e) => setPersonnelId(e.target.value)}
              placeholder={t('auth.personnelIdPlaceholder')}
              autoFocus
              disabled={isSubmitting}
            />
          </Field>

          <Field>
            <Label>{t('auth.pin')}</Label>
            <Input
              type="password"
              value={pin}
              onChange={(e) => setPin(e.target.value)}
              placeholder={t('auth.pinPlaceholder')}
              disabled={isSubmitting}
            />
          </Field>

          <Field>
            <Label>{t('auth.credentialKind')}</Label>
            <div
              className="sr-credential-options"
              role="radiogroup"
              aria-label={t('auth.credentialKind')}
            >
              {CREDENTIAL_KINDS.map((kind) => (
                <Button
                  key={kind}
                  type="button"
                  variant={credentialKind === kind ? 'primary' : 'secondary'}
                  onClick={() => handleCredentialKind(kind)}
                  disabled={isSubmitting || credentialStatus === 'waiting'}
                  aria-pressed={credentialKind === kind}
                >
                  {t(`auth.credentialKinds.${kind}`)}
                </Button>
              ))}
            </div>
          </Field>

          <div className="sr-field-row">
            <span className="sr-field-label">
              {credentialStatus === 'idle' && t('auth.credentialPrompt')}
              {credentialStatus === 'waiting' && t('auth.credentialWaiting')}
              {credentialStatus === 'detected' &&
                t('auth.credentialDetected', {
                  value: credentialRead?.maskedToken ?? t(`auth.credentialKinds.${credentialKind}`),
                })}
              {credentialStatus === 'error' && t('auth.credentialError')}
            </span>
            <Button
              type="button"
              variant="secondary"
              onClick={handleCredentialRead}
              disabled={isSubmitting || credentialStatus === 'waiting'}
            >
              {credentialStatus === 'detected'
                ? t('auth.rereadCredential')
                : t('auth.readCredential')}
            </Button>
          </div>

          {error && <p className="sr-field-error">{error}</p>}

          {attempts > 0 && attempts < MAX_ATTEMPTS && (
            <p className="sr-field-error">
              {t('auth.attemptsRemaining', { count: MAX_ATTEMPTS - attempts })}
            </p>
          )}

          <Button type="submit" disabled={isSubmitting || !personnelId || !pin || !credentialRead}>
            {isSubmitting ? t('auth.signingIn') : t('auth.signIn')}
          </Button>
        </form>
      </div>
    </div>
  );
}
