import { useCallback, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { useAuth } from '@/auth/AuthProvider.js';
import {
  readStaffCredential,
  type StaffCredentialKind,
  type StaffCredentialRead,
} from '@/auth/ibutton.js';
import { getStoreId } from '@/api-client-config.js';
import { getRecentLogins, recordLogin } from '@/lib/login-history.js';
import { reduceLoginWizardStep, type LoginWizardStep } from '@/lib/login-wizard.js';

import { CredentialStep } from './login/CredentialStep.js';
import { LoginLeftPanel } from './login/LoginLeftPanel.js';
import './login/LoginPage.css';
import { PersonnelIdStep } from './login/PersonnelIdStep.js';
import { PinStep } from './login/PinStep.js';

const MAX_ATTEMPTS = 5;
const ATTEMPTS_KEY = 'mercadia.sr-terminal.login-attempts';

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

/**
 * Three-step split-screen login wizard (plan 020), replacing the single-card
 * form. Layout/flow only — the underlying auth contract is unchanged:
 * `login()` (== `createAuthSession`) still validates personnel ID + PIN +
 * credential factor together in ONE atomic call
 * (`backend/services/store-edge/internal/app/auth.go`'s `CreateSession`),
 * fired exactly once here, automatically, right after step 3's credential
 * read succeeds (`attemptLogin`, called from `handleReadCredential`) — there
 * is no progressive/partial validation endpoint to stage this against, and
 * this file must not invent one.
 */
export function LoginPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { login, session } = useAuth();
  const storeId = useMemo(() => getStoreId(), []);

  const [step, setStep] = useState<LoginWizardStep>('personnelId');
  const [personnelId, setPersonnelId] = useState('');
  const [pin, setPin] = useState('');
  const [credentialKind, setCredentialKind] = useState<StaffCredentialKind>('ibutton');
  const [credentialRead, setCredentialRead] = useState<StaffCredentialRead | null>(null);
  const [credentialStatus, setCredentialStatus] = useState<
    'idle' | 'waiting' | 'detected' | 'error'
  >('idle');
  const [authError, setAuthError] = useState('');
  const [credentialError, setCredentialError] = useState('');
  const [attempts, setAttempts] = useState(loadAttempts);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [recentLogins] = useState(getRecentLogins);
  const isBlocked = attempts >= MAX_ATTEMPTS;

  const handleAdvance = useCallback(() => {
    setStep((current) => reduceLoginWizardStep(current, { type: 'advance' }));
  }, []);

  const handleSelectRecent = useCallback((actorId: string) => {
    setPersonnelId(actorId);
  }, []);

  /** Step 2's "Сменить" affordance: back to step 1, clearing only the PIN
   * and any step-3 credential state (personnelId itself is left as-is so
   * the operator can edit/confirm it rather than retyping from scratch). */
  const handleChangeIdentity = useCallback(() => {
    setStep((current) => reduceLoginWizardStep(current, { type: 'changeIdentity' }));
    setPin('');
    setCredentialRead(null);
    setCredentialStatus('idle');
    setCredentialError('');
    setAuthError('');
  }, []);

  /** Step 3's "Отменить вход" affordance: full reset back to step 1. Does
   * NOT touch `attempts` — the lockout counter tracks failed login()
   * submissions, not wizard navigation (plan 020). */
  const handleCancel = useCallback(() => {
    setStep((current) => reduceLoginWizardStep(current, { type: 'cancel' }));
    setPersonnelId('');
    setPin('');
    setCredentialKind('ibutton');
    setCredentialRead(null);
    setCredentialStatus('idle');
    setCredentialError('');
    setAuthError('');
  }, []);

  const handleCredentialKindChange = useCallback((kind: StaffCredentialKind) => {
    setCredentialKind(kind);
    setCredentialRead(null);
    setCredentialStatus('idle');
    setCredentialError('');
  }, []);

  /** The one real auth call, reusing `handleSubmit`'s exact former logic
   * (lockout bump on "Invalid credentials", error copy, role-based redirect)
   * — see this file's top doc comment. */
  const attemptLogin = useCallback(
    async (read: StaffCredentialRead) => {
      if (isBlocked || isSubmitting) return;

      setIsSubmitting(true);
      setAuthError('');
      setCredentialError('');

      try {
        const sess = await login(personnelId, pin, read.factor);
        recordLogin(sess.actorId);
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
        setAuthError(t('auth.invalidCredentials'));
      } finally {
        setIsSubmitting(false);
      }
    },
    [personnelId, pin, login, navigate, t, attempts, isSubmitting, isBlocked],
  );

  const handleReadCredential = useCallback(async () => {
    if (isBlocked || isSubmitting) return;

    setCredentialStatus('waiting');
    setCredentialError('');
    try {
      const nextRead = await readStaffCredential(credentialKind);
      setCredentialRead(nextRead);
      setCredentialStatus('detected');
      await attemptLogin(nextRead);
    } catch {
      setCredentialRead(null);
      setCredentialStatus('error');
      setCredentialError(t('auth.credentialError'));
    }
  }, [credentialKind, t, attemptLogin, isBlocked, isSubmitting]);

  if (session) {
    const target =
      session.roles.includes('senior_cashier') || session.roles.includes('admin')
        ? '/dashboard'
        : '/monitoring';
    navigate(target, { replace: true });
    return null;
  }

  return (
    <div className="sr-login-shell">
      <LoginLeftPanel currentStep={step} storeId={storeId} />

      <div className="sr-login-right">
        {step === 'personnelId' && (
          <PersonnelIdStep
            personnelId={personnelId}
            onPersonnelIdChange={setPersonnelId}
            onContinue={handleAdvance}
            recentLogins={recentLogins}
            onSelectRecent={handleSelectRecent}
          />
        )}

        {step === 'pin' && (
          <PinStep
            personnelId={personnelId}
            pin={pin}
            onPinChange={setPin}
            onEnter={handleAdvance}
            onChangeIdentity={handleChangeIdentity}
            maxAttempts={MAX_ATTEMPTS}
          />
        )}

        {step === 'credential' && (
          <CredentialStep
            personnelId={personnelId}
            credentialKind={credentialKind}
            onCredentialKindChange={handleCredentialKindChange}
            credentialStatus={credentialStatus}
            credentialRead={credentialRead}
            onReadCredential={() => void handleReadCredential()}
            isSubmitting={isSubmitting}
            isBlocked={isBlocked}
            attempts={attempts}
            maxAttempts={MAX_ATTEMPTS}
            authError={authError}
            credentialError={credentialError}
            onCancel={handleCancel}
          />
        )}
      </div>
    </div>
  );
}
