import { useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button, Input, Field, Label } from '@mercadia/ui';

import { useAuth, readIButton } from '@/auth/AuthProvider.js';
import { useIdleTimer } from '@/lib/use-idle-timer.js';

const MAX_ATTEMPTS = 5;

export function LoginPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { login, session } = useAuth();

  const [personnelId, setPersonnelId] = useState('');
  const [pin, setPin] = useState('');
  const [ibuttonStatus, setIbuttonStatus] = useState<'idle' | 'waiting' | 'detected' | 'error'>('idle');
  const [ibuttonRomId, setIbuttonRomId] = useState('');
  const [error, setError] = useState('');
  const [attempts, setAttempts] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useIdleTimer();

  const handleIButton = useCallback(async () => {
    setIbuttonStatus('waiting');
    setError('');
    try {
      const romId = await readIButton();
      setIbuttonRomId(romId);
      setIbuttonStatus('detected');
    } catch {
      setIbuttonStatus('error');
      setError(t('auth.ibuttonError'));
    }
  }, [t]);

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
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

    setIsSubmitting(true);
    setError('');

    try {
      await login(personnelId, pin, ibuttonRomId);
      navigate('/dashboard', { replace: true });
    } catch {
      setAttempts((prev) => prev + 1);
      setError(t('auth.invalidCredentials'));
    } finally {
      setIsSubmitting(false);
    }
  }, [personnelId, pin, ibuttonRomId, login, navigate, t, attempts, isSubmitting]);

  if (session) {
    navigate('/dashboard', { replace: true });
    return null;
  }

  return (
    <div className="sr-terminal-shell" style={{ alignItems: 'center', justifyContent: 'center' }}>
      <div className="sr-panel" style={{ width: 400, padding: '2rem' }}>
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

          <div className="sr-field-row">
            <span className="sr-field-label">
              {ibuttonStatus === 'idle' && t('auth.ibutton')}
              {ibuttonStatus === 'waiting' && t('auth.ibuttonWaiting')}
              {ibuttonStatus === 'detected' && t('auth.ibuttonDetected')}
              {ibuttonStatus === 'error' && t('auth.ibuttonError')}
            </span>
            <Button
              type="button"
              variant="secondary"
              onClick={handleIButton}
              disabled={isSubmitting || ibuttonStatus === 'waiting'}
            >
              {ibuttonStatus === 'detected' ? '✓' : t('ibutton.present')}
            </Button>
          </div>

          {error && <p className="sr-field-error">{error}</p>}

          {attempts > 0 && attempts < MAX_ATTEMPTS && (
            <p className="sr-field-error">
              {t('auth.attemptsRemaining', { count: MAX_ATTEMPTS - attempts })}
            </p>
          )}

          <Button type="submit" disabled={isSubmitting || !personnelId || !pin}>
            {isSubmitting ? t('auth.signingIn') : t('auth.signIn')}
          </Button>
        </form>
      </div>
    </div>
  );
}
