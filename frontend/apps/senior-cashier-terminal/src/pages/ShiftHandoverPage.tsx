import { useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button, Input, Field, Label } from '@mercadia/ui';

import { useAuth, readIButton } from '@/auth/AuthProvider.js';
import { TerminalHeader } from '@/components/TerminalHeader.js';

export function ShiftHandoverPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { session, logout } = useAuth();

  const [incomingId, setIncomingId] = useState('');
  const [incomingPin, setIncomingPin] = useState('');
  const [ibuttonStatus, setIbuttonStatus] = useState<'idle' | 'waiting' | 'detected' | 'error'>('idle');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleIButton = useCallback(async () => {
    setIbuttonStatus('waiting');
    setError('');
    try {
      await readIButton();
      setIbuttonStatus('detected');
    } catch {
      setIbuttonStatus('error');
      setError(t('ibutton.error'));
    }
  }, [t]);

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      if (isSubmitting) return;

      if (!incomingId || !incomingPin) {
        setError(t('auth.invalidCredentials'));
        return;
      }

      setIsSubmitting(true);
      setError('');

      try {
        // In a real flow, we would:
        // 1. Authenticate the incoming senior cashier via Store Edge
        // 2. Close current session
        // 3. Open new session
        // For now, just logout current user and redirect to login
        logout();
        navigate('/login', { replace: true });
      } catch {
        setError(t('common.unexpectedError'));
      } finally {
        setIsSubmitting(false);
      }
    },
    [incomingId, incomingPin, logout, navigate, t, isSubmitting],
  );

  return (
    <div className="sr-terminal-shell">
      <TerminalHeader title={t('handover.title')} onLogout={() => navigate('/login')} />

      <main className="sr-terminal-main">
        <form onSubmit={handleSubmit} className="sr-form">
          <div className="sr-panel">
            <h3 className="sr-panel-title">{t('handover.currentSession')}</h3>
            <p className="muted">
              {session?.actorId} · {session?.roles?.join(', ')}
            </p>
          </div>

          <div className="sr-panel">
            <h3 className="sr-panel-title">{t('handover.incomingCashier')}</h3>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', marginTop: '0.75rem' }}>
              <Field>
                <Label>{t('handover.incomingPersonnelId')}</Label>
                <Input
                  value={incomingId}
                  onChange={(e) => setIncomingId(e.target.value)}
                  disabled={isSubmitting}
                />
              </Field>

              <Field>
                <Label>{t('handover.incomingPin')}</Label>
                <Input
                  type="password"
                  value={incomingPin}
                  onChange={(e) => setIncomingPin(e.target.value)}
                  disabled={isSubmitting}
                />
              </Field>

              <div className="sr-field-row">
                <span className="sr-field-label">
                  {ibuttonStatus === 'idle' && t('handover.incomingIbutton')}
                  {ibuttonStatus === 'waiting' && t('ibutton.waiting')}
                  {ibuttonStatus === 'detected' && t('ibutton.detected')}
                  {ibuttonStatus === 'error' && t('ibutton.error')}
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
            </div>
          </div>

          {error && <p className="sr-field-error">{error}</p>}

          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <Button type="button" variant="ghost" onClick={() => navigate('/dashboard')}>
              {t('common.cancel')}
            </Button>
            <Button type="submit" disabled={isSubmitting || !incomingId || !incomingPin}>
              {isSubmitting ? t('common.submitting') : t('handover.confirmHandover')}
            </Button>
          </div>
        </form>
      </main>
    </div>
  );
}
