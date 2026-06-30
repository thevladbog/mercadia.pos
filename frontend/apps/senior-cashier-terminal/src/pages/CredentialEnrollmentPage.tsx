import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button, Field, Label, Select } from '@mercadia/ui';
import {
  addActorCredentialBinding,
  useGetCredentialManagement,
  type GetCredentialManagement200ActorsItem,
} from '@mercadia/api-clients-store-edge';

import { useAuth } from '@/auth/AuthProvider.js';
import {
  readStaffCredential,
  type StaffCredentialKind,
  type StaffCredentialRead,
} from '@/auth/ibutton.js';
import { TerminalHeader } from '@/components/TerminalHeader.js';
import { getStoreId } from '@/api-client-config.js';
import { useIdleTimer } from '@/lib/use-idle-timer.js';

const CREDENTIAL_KINDS: StaffCredentialKind[] = ['ibutton', 'msr_card', 'barcode_card'];

function actorLabel(actor: GetCredentialManagement200ActorsItem): string {
  return `${actor.id} (${actor.roles.join(', ')})`;
}

export function CredentialEnrollmentPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { session, logout } = useAuth();
  const storeId = useMemo(() => getStoreId(), []);
  const { remaining } = useIdleTimer();
  const [selectedActorId, setSelectedActorId] = useState('');
  const [credentialKind, setCredentialKind] = useState<StaffCredentialKind>('ibutton');
  const [credentialRead, setCredentialRead] = useState<StaffCredentialRead | null>(null);
  const [isReading, setIsReading] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');
  const readAbortControllerRef = useRef<AbortController | null>(null);

  const credentialsQuery = useGetCredentialManagement(storeId);
  const credentials = credentialsQuery.data?.status === 200 ? credentialsQuery.data.data : null;
  const actors = credentials?.actors ?? [];
  const selectedActor = actors.find((actor) => actor.id === selectedActorId) ?? actors[0] ?? null;
  const targetActorId = selectedActor?.id ?? '';
  const isSelfTarget = Boolean(session?.actorId && targetActorId === session.actorId);

  useEffect(() => {
    return () => {
      readAbortControllerRef.current?.abort();
      readAbortControllerRef.current = null;
    };
  }, []);

  const readCredential = useCallback(async () => {
    readAbortControllerRef.current?.abort();
    const abortController = new AbortController();
    readAbortControllerRef.current = abortController;
    setIsReading(true);
    setMessage('');
    setError('');
    try {
      const nextRead = await readStaffCredential(credentialKind, abortController.signal);
      if (abortController.signal.aborted) return;
      setCredentialRead(nextRead);
      setMessage(
        t('credentials.readSuccess', {
          value: nextRead.maskedToken ?? t(`auth.credentialKinds.${credentialKind}`),
        }),
      );
    } catch {
      if (abortController.signal.aborted) return;
      setCredentialRead(null);
      setError(t('credentials.readError'));
    } finally {
      if (readAbortControllerRef.current === abortController) {
        readAbortControllerRef.current = null;
        if (!abortController.signal.aborted) {
          setIsReading(false);
        }
      }
    }
  }, [credentialKind, t]);

  const handleBack = useCallback(() => {
    readAbortControllerRef.current?.abort();
    navigate('/dashboard');
  }, [navigate]);

  const submitEnrollment = useCallback(async () => {
    if (!selectedActor || !credentialRead || isSelfTarget) return;

    setIsSubmitting(true);
    setMessage('');
    setError('');
    try {
      const response = await addActorCredentialBinding(
        storeId,
        selectedActor.id,
        {
          kind: credentialRead.factor.kind,
          token: credentialRead.factor.token,
          maskedToken: credentialRead.maskedToken,
        },
        { headers: { 'Idempotency-Key': crypto.randomUUID() } },
      );
      if (response.status === 200) {
        setCredentialRead(null);
        await credentialsQuery.refetch();
        setMessage(t('credentials.enrollSuccess', { actorId: selectedActor.id }));
      }
    } catch {
      setError(t('credentials.enrollError'));
    } finally {
      setIsSubmitting(false);
    }
  }, [credentialRead, credentialsQuery, isSelfTarget, selectedActor, storeId, t]);

  const formatRemaining = (ms: number) => {
    const totalSec = Math.floor(ms / 1000);
    const h = Math.floor(totalSec / 3600);
    const m = Math.floor((totalSec % 3600) / 60);
    return `${h}${t('dashboard.hours')} ${m}${t('dashboard.minutes')}`;
  };

  return (
    <div className="sr-terminal-shell">
      <TerminalHeader title={t('credentials.title')} onLogout={logout} />

      <main className="sr-terminal-main">
        <div className="sr-panel sr-stack">
          <div className="sr-panel-header">
            <div>
              <h2 className="sr-panel-title">{t('credentials.enrollTitle')}</h2>
              <p className="muted">{t('credentials.subtitle')}</p>
            </div>
            <span className="muted" style={{ fontSize: '0.85rem' }}>
              {t('dashboard.autoLockIn')}: {formatRemaining(remaining)}
            </span>
          </div>

          <div className="sr-form sr-form-wide">
            <Field>
              <Label>{t('credentials.employee')}</Label>
              <Select
                value={targetActorId}
                onChange={(event) => {
                  setSelectedActorId(event.target.value);
                  setCredentialRead(null);
                  setMessage('');
                  setError('');
                }}
                disabled={isReading || isSubmitting || actors.length === 0}
              >
                {actors.map((actor) => (
                  <option key={actor.id} value={actor.id}>
                    {actorLabel(actor)}
                  </option>
                ))}
              </Select>
            </Field>

            <Field>
              <Label>{t('credentials.kind')}</Label>
              <div
                className="sr-credential-options"
                role="group"
                aria-label={t('credentials.kind')}
              >
                {CREDENTIAL_KINDS.map((kind) => (
                  <Button
                    key={kind}
                    type="button"
                    variant={credentialKind === kind ? 'primary' : 'secondary'}
                    onClick={() => {
                      setCredentialKind(kind);
                      setCredentialRead(null);
                      setMessage('');
                      setError('');
                    }}
                    disabled={isReading || isSubmitting}
                    aria-pressed={credentialKind === kind}
                  >
                    {t(`auth.credentialKinds.${kind}`)}
                  </Button>
                ))}
              </div>
            </Field>

            <div className="sr-field-row">
              <span className="sr-field-label">
                {credentialRead
                  ? t('credentials.readSuccess', {
                      value:
                        credentialRead.maskedToken ?? t(`auth.credentialKinds.${credentialKind}`),
                    })
                  : t('credentials.readPrompt')}
              </span>
              <Button
                type="button"
                variant="secondary"
                onClick={() => void readCredential()}
                disabled={isReading || isSubmitting || !selectedActor}
              >
                {isReading
                  ? t('credentials.reading')
                  : credentialRead
                    ? t('credentials.reread')
                    : t('credentials.read')}
              </Button>
            </div>

            {isSelfTarget && <p className="sr-field-error">{t('credentials.selfTarget')}</p>}
            {credentialsQuery.isLoading && <p className="muted">{t('common.loading')}</p>}
            {credentialsQuery.isError && (
              <p className="sr-field-error">{t('credentials.loadError')}</p>
            )}
            {actors.length === 0 && !credentialsQuery.isLoading && (
              <p className="muted">{t('credentials.noActors')}</p>
            )}
            {message && <p className="success-text">{message}</p>}
            {error && <p className="sr-field-error">{error}</p>}

            <div className="sr-button-row">
              <Button type="button" variant="secondary" onClick={handleBack}>
                {t('common.back')}
              </Button>
              <Button
                type="button"
                onClick={() => void submitEnrollment()}
                disabled={
                  isSubmitting || isReading || !credentialRead || !selectedActor || isSelfTarget
                }
              >
                {isSubmitting ? t('common.submitting') : t('credentials.enroll')}
              </Button>
            </div>
          </div>
        </div>

        {selectedActor && (
          <div className="sr-panel sr-stack" style={{ marginTop: '1rem' }}>
            <h2 className="sr-panel-title">{t('credentials.currentBindings')}</h2>
            {selectedActor.credentialBindings.length === 0 ? (
              <p className="muted">{t('credentials.noBindings')}</p>
            ) : (
              <div className="sr-binding-list">
                {selectedActor.credentialBindings.map((binding) => (
                  <div
                    key={`${binding.kind}-${binding.tokenFingerprint}`}
                    className="sr-binding-card"
                  >
                    <div>
                      <strong>{t(`auth.credentialKinds.${binding.kind}`)}</strong>
                      <div className="muted">{binding.maskedToken || binding.tokenFingerprint}</div>
                    </div>
                    <span className={binding.active ? 'success-text' : 'muted'}>
                      {binding.active ? t('credentials.active') : t('credentials.revoked')}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </main>
    </div>
  );
}
