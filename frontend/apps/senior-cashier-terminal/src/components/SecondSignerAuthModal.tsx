import { useCallback, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ApiError, createAuthSession } from '@mercadia/api-clients-store-edge';
import { Dialog, DialogBody, DialogContent, DialogTitle } from '@mercadia/ui';

import { getStoreId } from '@/api-client-config.js';
import { useAuth } from '@/auth/AuthProvider.js';
import {
  readStaffCredential,
  type StaffCredentialKind,
  type StaffCredentialRead,
} from '@/auth/ibutton.js';
import { actorsMustDiffer } from '@/lib/cash-utils.js';
import { getRecentLogins, recordLogin } from '@/lib/login-history.js';
import { reduceLoginWizardStep, type LoginWizardStep } from '@/lib/login-wizard.js';

import { CredentialStep } from '../pages/login/CredentialStep.js';
import '../pages/login/LoginPage.css';
import { PersonnelIdStep } from '../pages/login/PersonnelIdStep.js';
import { PinStep } from '../pages/login/PinStep.js';

const MAX_ATTEMPTS = 5;

export interface SecondSignerAuthModalProps {
  open: boolean;
  onClose: () => void;
  /** Called once a DIFFERENT actor than the primary signed-in session
   * successfully authenticates. Never fired for the same actorId — see the
   * client-side guard in `attemptAuth` below. */
  onAuthenticated: (actorId: string) => void;
}

/**
 * Re-authentication modal for a second senior cashier/admin to sign off a
 * safe-recount discrepancy (plan 028, design screen `08b`). Reuses the exact
 * same 3-step wizard as `LoginPage.tsx` (`PersonnelIdStep` → `PinStep` →
 * `CredentialStep`, driven by the same pure `login-wizard.ts` state machine)
 * — all three step components are pure/prop-driven with zero session or
 * routing dependencies, confirmed by reading each file in full, so they're
 * reusable here verbatim with entirely local state instead of `LoginPage`'s
 * global session state. Wrapped in `Dialog`/`DialogContent` (the same shell
 * `MismatchDialog.tsx` already uses) instead of `LoginLeftPanel`'s full-page
 * split-screen layout — `LoginPage.css`'s classes are still reused as-is,
 * since none of them depend on being a direct child of `.sr-login-right`.
 *
 * CRITICAL — this must call the raw generated `createAuthSession` function
 * DIRECTLY (same import `AuthProvider.tsx` itself uses), never
 * `useAuth().login()`. `login()` calls `saveSession`/`setSession`/
 * `setSessionToken`, all of which overwrite the ONE global session and
 * Authorization header used by the whole app/every subsequent API call —
 * using it here would silently log out the primary signed-in senior
 * cashier. The second signer's session token is therefore never saved
 * anywhere; only their `actorId` (once confirmed) is handed back via
 * `onAuthenticated`.
 *
 * Keeps its OWN local attempt counter — reset every time the modal is
 * (re)opened — deliberately NOT shared with `LoginPage.tsx`'s
 * `ATTEMPTS_KEY`/`sessionStorage`-backed counter: a different terminal
 * login flow's lockout must not be affected by second-signer auth attempts.
 */
export function SecondSignerAuthModal({ open, onClose, onAuthenticated }: SecondSignerAuthModalProps) {
  const { t } = useTranslation();
  const { session } = useAuth();

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
  const [attempts, setAttempts] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [recentLogins, setRecentLogins] = useState(getRecentLogins);
  const isBlocked = attempts >= MAX_ATTEMPTS;

  // Reset the whole wizard every time the dialog transitions from closed to
  // open, so a previous second-signer attempt (or lockout) never leaks into
  // the next one. Adjusts state during render (React's documented pattern
  // for "resetting state when a prop changes") rather than in a `useEffect`,
  // which would fire a redundant extra render after the one that flipped
  // `open` — see https://react.dev/learn/you-might-not-need-an-effect.
  const [wasOpen, setWasOpen] = useState(open);
  if (open && !wasOpen) {
    setWasOpen(true);
    setStep('personnelId');
    setPersonnelId('');
    setPin('');
    setCredentialKind('ibutton');
    setCredentialRead(null);
    setCredentialStatus('idle');
    setAuthError('');
    setCredentialError('');
    setAttempts(0);
    setIsSubmitting(false);
    setRecentLogins(getRecentLogins());
  } else if (!open && wasOpen) {
    setWasOpen(false);
  }

  const handleAdvance = useCallback(() => {
    setStep((current) => reduceLoginWizardStep(current, { type: 'advance' }));
  }, []);

  const handleSelectRecent = useCallback((actorId: string) => {
    setPersonnelId(actorId);
  }, []);

  const handleChangeIdentity = useCallback(() => {
    setStep((current) => reduceLoginWizardStep(current, { type: 'changeIdentity' }));
    setPin('');
    setCredentialRead(null);
    setCredentialStatus('idle');
    setCredentialError('');
    setAuthError('');
  }, []);

  const handleCancelStep = useCallback(() => {
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

  const attemptAuth = useCallback(
    async (read: StaffCredentialRead) => {
      if (isBlocked || isSubmitting) return;

      setIsSubmitting(true);
      setAuthError('');
      setCredentialError('');

      try {
        const response = await createAuthSession({
          actorId: personnelId,
          pin,
          storeId: getStoreId(),
          credentialFactor: read.factor,
        });
        if (response.status !== 201) {
          throw new Error('Authentication failed');
        }

        // The token is deliberately discarded here — never passed to
        // `setSessionToken`/`saveSession` — only the confirmed actorId is
        // kept. See this file's top doc comment.
        const authenticatedActorId = response.data.session.actorId;

        if (!actorsMustDiffer(authenticatedActorId, session?.actorId ?? '')) {
          setAuthError(t('cash.secondSigner.sameActorError'));
          return;
        }

        recordLogin(authenticatedActorId);
        onAuthenticated(authenticatedActorId);
      } catch (err) {
        if (err instanceof ApiError && err.status === 401) {
          setAttempts((current) => current + 1);
        }
        setAuthError(t('auth.invalidCredentials'));
      } finally {
        setIsSubmitting(false);
      }
    },
    [personnelId, pin, session, onAuthenticated, t, isSubmitting, isBlocked],
  );

  const handleReadCredential = useCallback(async () => {
    if (isBlocked || isSubmitting) return;

    setCredentialStatus('waiting');
    setCredentialError('');
    try {
      const nextRead = await readStaffCredential(credentialKind);
      setCredentialRead(nextRead);
      setCredentialStatus('detected');
      await attemptAuth(nextRead);
    } catch {
      setCredentialRead(null);
      setCredentialStatus('error');
      setCredentialError(t('auth.credentialError'));
    }
  }, [credentialKind, t, attemptAuth, isBlocked, isSubmitting]);

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next) onClose();
      }}
    >
      <DialogContent aria-describedby={undefined}>
        <DialogTitle>{t('cash.secondSigner.title')}</DialogTitle>
        <DialogBody>
          <p className="muted">
            {t('cash.secondSigner.primaryStaysAuthorized', { actorId: session?.actorId ?? '' })}
          </p>

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
              onCancel={handleCancelStep}
            />
          )}
        </DialogBody>
      </DialogContent>
    </Dialog>
  );
}
