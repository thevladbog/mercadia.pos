import { useCallback, useMemo, useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { AvatarChip, Badge, Button } from '@mercadia/ui';
import {
  useGetCredentialManagement,
  useListCashMovements,
  useListCashRecounts,
  useListOperationJournal,
  useListStoreChangeFundRequests,
  useListStoreReturns,
} from '@mercadia/api-clients-store-edge';

import { useAuth } from '@/auth/AuthProvider.js';
import {
  readStaffCredential,
  type StaffCredentialKind,
  type StaffCredentialRead,
} from '@/auth/ibutton.js';
import { getStoreId } from '@/api-client-config.js';
import { formatMinor, selectSuccessData } from '@/lib/cash-utils.js';
import {
  deriveOperationsCount,
  type CredentialActorForJoin,
  type OperationJournalEntryForJoin,
} from '@/lib/dashboard-data.js';
import {
  deriveHandoverSummary,
  deriveHandoverWarnings,
  joinEligibleSuccessors,
  type CashMovementForJoin,
  type CashRecountForJoin,
  type ChangeFundRequestForJoin,
  type EligibleSuccessor,
  type ReturnForJoin,
} from '@/lib/handover-data.js';
import { reduceHandoverWizardStep, type HandoverWizardStep } from '@/lib/handover-wizard.js';
import {
  deriveElapsed,
  deriveInitials,
  formatLoginTime,
  formatRoleLabel,
} from '@/lib/topbar-utils.js';
import { useTopBarActions } from '@/lib/use-topbar-actions.js';
import { TopBar } from '@/components/TopBar.js';

import { CredentialStep } from './login/CredentialStep.js';
import './login/LoginPage.css';
import { PinStep } from './login/PinStep.js';
import './ShiftHandoverPage.css';

const MAX_ATTEMPTS = 5;

// The API-enforced max for every list endpoint this page reads (`limit` on
// `ListCashMovementsParams`/`ListStoreReturnsParams`/
// `ListStoreChangeFundRequestsParams`/`ListCashRecountsParams`/
// `ListOperationJournalParams`) — same "real-data approximation, not a true
// bounded count" caveat `dashboard-data.ts`'s `deriveOperationsCount` already
// documents for `JOURNAL_FETCH_LIMIT`.
const SUMMARY_FETCH_LIMIT = 100;

// Stable empty-array fallbacks (module scope, never reassigned) — same
// convention as `dashboard-data.ts`'s `EMPTY_SHIFTS`/`EMPTY_TERMINALS`/
// `EMPTY_ACTORS` (plan 021).
const EMPTY_ACTORS: CredentialActorForJoin[] = [];
const EMPTY_MOVEMENTS: CashMovementForJoin[] = [];
const EMPTY_RETURNS: ReturnForJoin[] = [];
const EMPTY_CHANGE_FUND_REQUESTS: ChangeFundRequestForJoin[] = [];
const EMPTY_CASH_RECOUNTS: CashRecountForJoin[] = [];
const EMPTY_JOURNAL: OperationJournalEntryForJoin[] = [];

/**
 * Redesigned Shift Handover page (plan 029, Phase 9 — the last phase of the
 * senior-cashier-terminal redesign initiative). Replaces the old stub (144
 * lines, manual `incomingId`/`incomingPin` text fields + a legacy
 * `readIButton()` call whose result was read and immediately discarded,
 * whose own submit-handler comment read "For now, just logout current user
 * and redirect to login") with a real handoff: a session-summary recap for
 * the closing senior cashier (`lib/handover-data.ts`'s `deriveHandoverSummary`/
 * `deriveHandoverWarnings`), a real successor picker
 * (`joinEligibleSuccessors`), and a genuine re-authentication step reusing
 * the exact same `PinStep`/`CredentialStep` components `SecondSignerAuthModal.tsx`
 * (plan 028) already established as pure/prop-driven/reusable — NOT
 * `PersonnelIdStep`, since the successor's identity is already known from
 * the picker selection, not typed (see `lib/handover-wizard.ts`'s 3-value
 * step type, which deliberately does not reuse `LoginWizardStep`'s type for
 * this reason).
 *
 * CRITICAL, easy to get backwards: `attemptHandover` below calls
 * `useAuth().login()` DIRECTLY — the OPPOSITE of what
 * `SecondSignerAuthModal.tsx` does. That modal deliberately calls the raw
 * `createAuthSession` and DISCARDS the token, because it is authenticating a
 * SECOND signer IN ADDITION to the primary session, which must not be
 * disturbed. A handoff is the opposite case: the successor's session SHOULD
 * become the new primary session — exactly what the ordinary `login()` call
 * already does (an entirely normal fresh login, no special handling
 * needed). Getting this backwards (copying `SecondSignerAuthModal`'s
 * discard-the-token approach here) would silently break the whole feature:
 * the successor's credentials would validate, but the terminal would never
 * actually change hands.
 *
 * Per an explicit user decision (2026-07-10, plan 029), no "Just lock"
 * button is built — every existing "lock" affordance in this app
 * (`useTopBarActions().onLock`) is already a full logout under a relabeled
 * name, and there is no server-side concept of "who is currently logged
 * into this terminal" to build a lock-without-logout mechanism around
 * (`SessionRepository` only has `SaveSession`/`FindSessionByToken` — no
 * list/delete/revoke). This page therefore ships with exactly 2 actions:
 * `Cancel` (unchanged, navigates to `/dashboard`) and `Hand off`. The
 * `Hand off` footer button is only meaningful on the `pickSuccessor` step —
 * picking a row there both selects AND advances the wizard directly (per
 * plan 029 scope item 4's literal "selecting one advances the wizard to
 * 'pin'"), so the footer button stays disabled on that step (there is
 * nothing left to confirm once a row is clicked) and is not rendered at all
 * on the `pin`/`credential` steps, which instead use `PinStep`/
 * `CredentialStep`'s own built-in navigation (`onChangeIdentity`/`onCancel`),
 * exactly mirroring `SecondSignerAuthModal.tsx`'s precedent of never
 * wrapping those components in extra page-level chrome.
 */
export function ShiftHandoverPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { onHandover } = useTopBarActions();
  const { session, loggedInAt, login } = useAuth();
  const storeId = useMemo(() => getStoreId(), []);

  const actorId = session?.actorId ?? '';
  // Falls back to the Unix epoch (i.e. "since the beginning of time," never
  // excluding anything) in the defensive edge case where `loggedInAt` is
  // somehow unavailable for an otherwise-valid session — this NEVER
  // reintroduces the fragile `expiresAt`-minus-assumed-TTL derivation this
  // plan explicitly avoids (see `AuthProvider.tsx`'s doc comment).
  const sinceIso = loggedInAt ?? new Date(0).toISOString();

  const { data: credentialsResp } = useGetCredentialManagement(storeId);
  const { data: movementsResp } = useListCashMovements(storeId, { limit: SUMMARY_FETCH_LIMIT });
  const { data: returnsResp } = useListStoreReturns(storeId, { limit: SUMMARY_FETCH_LIMIT });
  const { data: changeFundRequestsResp } = useListStoreChangeFundRequests(storeId, {
    limit: SUMMARY_FETCH_LIMIT,
  });
  const { data: cashRecountsResp } = useListCashRecounts(storeId, { limit: SUMMARY_FETCH_LIMIT });
  const { data: journalResp } = useListOperationJournal(storeId, { limit: SUMMARY_FETCH_LIMIT });

  const actors =
    selectSuccessData<{ actors: CredentialActorForJoin[] }>(credentialsResp)?.actors ??
    EMPTY_ACTORS;
  const movements =
    selectSuccessData<{ items: CashMovementForJoin[] }>(movementsResp)?.items ?? EMPTY_MOVEMENTS;
  const returns =
    selectSuccessData<{ items: ReturnForJoin[] }>(returnsResp)?.items ?? EMPTY_RETURNS;
  const changeFundRequests =
    selectSuccessData<{ items: ChangeFundRequestForJoin[] }>(changeFundRequestsResp)?.items ??
    EMPTY_CHANGE_FUND_REQUESTS;
  const cashRecounts =
    selectSuccessData<{ items: CashRecountForJoin[] }>(cashRecountsResp)?.items ??
    EMPTY_CASH_RECOUNTS;
  const journalItems =
    selectSuccessData<{ items: OperationJournalEntryForJoin[] }>(journalResp)?.items ??
    EMPTY_JOURNAL;

  const successors = useMemo(() => joinEligibleSuccessors(actors, actorId), [actors, actorId]);
  const summary = useMemo(
    () => deriveHandoverSummary(movements, returns, actorId, sinceIso),
    [movements, returns, actorId, sinceIso],
  );
  const warnings = useMemo(
    () => deriveHandoverWarnings(changeFundRequests, cashRecounts),
    [changeFundRequests, cashRecounts],
  );
  // The single overall "N operations this session" figure (design's "14") —
  // reuses `deriveOperationsCount` verbatim, distinct from each summary
  // row's own per-category count above.
  const operationsCount = useMemo(
    () => deriveOperationsCount(journalItems, actorId, sinceIso),
    [journalItems, actorId, sinceIso],
  );

  const [step, setStep] = useState<HandoverWizardStep>('pickSuccessor');
  const [selectedSuccessor, setSelectedSuccessor] = useState<EligibleSuccessor | null>(null);
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
  const isBlocked = attempts >= MAX_ATTEMPTS;

  // Ticking "now" for the elapsed-time line below, same pattern
  // `TopBar.tsx` already uses for its own "since HH:MM (XhYm)" display —
  // `Date.now()` must not be called directly in the render body (React's
  // purity rule), only inside a lazy `useState` initializer or an effect.
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(id);
  }, []);

  /** Picking a candidate both selects them and advances the wizard, per
   * plan 029 scope item 4 ("selecting one advances the wizard to 'pin'") —
   * there is no separate confirm step. */
  const handleSelectSuccessor = useCallback((successor: EligibleSuccessor) => {
    setSelectedSuccessor(successor);
    setStep((current) => reduceHandoverWizardStep(current, { type: 'advance' }));
  }, []);

  const handleAdvance = useCallback(() => {
    setStep((current) => reduceHandoverWizardStep(current, { type: 'advance' }));
  }, []);

  /** `PinStep`'s "Сменить" affordance: back to the picker, clearing the PIN
   * and any credential-step state (mirrors `LoginPage.tsx`'s
   * `handleChangeIdentity`). */
  const handleChangeSuccessor = useCallback(() => {
    setStep((current) => reduceHandoverWizardStep(current, { type: 'changeSuccessor' }));
    setSelectedSuccessor(null);
    setPin('');
    setCredentialRead(null);
    setCredentialStatus('idle');
    setCredentialError('');
    setAuthError('');
  }, []);

  /** `CredentialStep`'s "Отменить" affordance: full reset back to the
   * picker. Does NOT touch `attempts` — the lockout counter tracks failed
   * `login()` submissions, not wizard navigation (mirrors `LoginPage.tsx`'s
   * `handleCancel`). */
  const handleCancelStep = useCallback(() => {
    setStep((current) => reduceHandoverWizardStep(current, { type: 'cancel' }));
    setSelectedSuccessor(null);
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

  /** The one real re-authentication call — see this file's top doc comment
   * on why this MUST be `login()`, never `createAuthSession`. */
  const attemptHandover = useCallback(
    async (read: StaffCredentialRead) => {
      if (isBlocked || isSubmitting || !selectedSuccessor) return;

      setIsSubmitting(true);
      setAuthError('');
      setCredentialError('');

      try {
        const nextSession = await login(selectedSuccessor.actorId, pin, read.factor);
        const target =
          nextSession.roles.includes('senior_cashier') || nextSession.roles.includes('admin')
            ? '/dashboard'
            : '/monitoring';
        navigate(target, { replace: true });
      } catch (err) {
        if (err instanceof Error && err.message === 'Invalid credentials') {
          setAttempts((current) => current + 1);
        }
        setAuthError(t('auth.invalidCredentials'));
      } finally {
        setIsSubmitting(false);
      }
    },
    [selectedSuccessor, pin, login, navigate, t, isBlocked, isSubmitting],
  );

  const handleReadCredential = useCallback(async () => {
    if (isBlocked || isSubmitting) return;

    setCredentialStatus('waiting');
    setCredentialError('');
    try {
      const nextRead = await readStaffCredential(credentialKind);
      setCredentialRead(nextRead);
      setCredentialStatus('detected');
      await attemptHandover(nextRead);
    } catch {
      setCredentialRead(null);
      setCredentialStatus('error');
      setCredentialError(t('auth.credentialError'));
    }
  }, [credentialKind, t, attemptHandover, isBlocked, isSubmitting]);

  const role = session?.roles[0] ? formatRoleLabel(session.roles[0], t) : '';
  const initials = deriveInitials(actorId);
  const loginAtMs = new Date(sinceIso).getTime();
  const elapsed = deriveElapsed(loginAtMs, now);

  const summaryRows: { key: string; label: string; amountMinor: number; count: number }[] = [
    {
      key: 'changeIssued',
      label: t('handover.summary.changeIssued'),
      amountMinor: summary.changeIssuedMinor,
      count: summary.changeIssuedCount,
    },
    {
      key: 'cashReceived',
      label: t('handover.summary.cashReceived'),
      amountMinor: summary.cashReceivedMinor,
      count: summary.cashReceivedCount,
    },
    {
      key: 'incassations',
      label: t('handover.summary.incassations'),
      amountMinor: summary.incassationsMinor,
      count: summary.incassationsCount,
    },
    {
      key: 'expenses',
      label: t('handover.summary.expenses'),
      amountMinor: summary.expensesMinor,
      count: summary.expensesCount,
    },
    {
      key: 'refundsApproved',
      label: t('handover.summary.refundsApproved'),
      amountMinor: summary.refundsApprovedMinor,
      count: summary.refundsApprovedCount,
    },
    {
      key: 'safeNetChange',
      label: t('handover.summary.safeNetChange'),
      amountMinor: summary.safeNetChangeMinor,
      count: summary.safeNetChangeCount,
    },
  ];

  const hasWarnings =
    warnings.pendingChangeFundRequests > 0 || warnings.openRecountDiscrepancies > 0;

  return (
    <div className="sr-terminal-shell">
      {/* `onLock` intentionally does NOT call `logout()` here (unlike
       * `useTopBarActions().onLock`) — see `use-topbar-actions.ts`'s own
       * doc comment: this page IS the handover flow itself, so its lock
       * button is deliberately excluded from that shared wiring. Preserved
       * unchanged from the pre-plan-029 page. */}
      <TopBar
        onHandover={onHandover}
        onLock={() => navigate('/login')}
        operationsCount={operationsCount}
      />

      <main className="sr-terminal-main">
        <h1 className="sr-page-title">{t('handover.title')}</h1>

        <div className="sr-handover-body">
          <div className="sr-handover-left">
            <div className="sr-panel">
              <div className="sr-handover-identity">
                <AvatarChip initials={initials} size="md" />
                <div className="sr-handover-identity-text">
                  <span className="sr-handover-identity-line">
                    {actorId} · {role}
                  </span>
                  <span className="muted">
                    {t('topbar.since', { time: formatLoginTime(loginAtMs) })} ({elapsed.hours}
                    {t('dashboard.hours')} {elapsed.minutes}
                    {t('dashboard.minutes')})
                  </span>
                </div>
              </div>

              <h2 className="sr-panel-title" style={{ marginTop: '1.25rem' }}>
                {t('handover.summary.title')}
              </h2>
              <div className="sr-handover-summary-list">
                {summaryRows.map((row) => (
                  <div className="sr-handover-summary-row" key={row.key}>
                    <div className="sr-handover-summary-label">
                      <span>{row.label}</span>
                      <span className="sr-handover-summary-sublabel">
                        {t('handover.summary.opCount', { count: row.count })}
                      </span>
                    </div>
                    <span className="sr-handover-summary-amount">
                      {formatMinor(row.amountMinor)} ₽
                    </span>
                  </div>
                ))}
              </div>
            </div>

            {hasWarnings && (
              <div className="sr-handover-warning" role="status">
                {warnings.pendingChangeFundRequests > 0 && (
                  <span>
                    {t('handover.warning.pendingChangeFundRequests', {
                      count: warnings.pendingChangeFundRequests,
                    })}
                  </span>
                )}
                {warnings.openRecountDiscrepancies > 0 && (
                  <span>
                    {t('handover.warning.openRecountDiscrepancies', {
                      count: warnings.openRecountDiscrepancies,
                    })}
                  </span>
                )}
              </div>
            )}
          </div>

          <div className="sr-handover-right">
            {step === 'pickSuccessor' && (
              <div className="sr-panel">
                <h2 className="sr-panel-title">{t('handover.pickSuccessor.title')}</h2>
                {!credentialsResp && <p className="muted">{t('common.loading')}</p>}
                {credentialsResp && successors.length === 0 && (
                  <p className="muted">{t('handover.pickSuccessor.empty')}</p>
                )}
                <div className="sr-handover-successor-list">
                  {successors.map((successor) => (
                    <button
                      key={successor.actorId}
                      type="button"
                      className="sr-handover-successor-row"
                      onClick={() => handleSelectSuccessor(successor)}
                    >
                      <AvatarChip initials={deriveInitials(successor.actorId)} size="sm" />
                      <span className="sr-handover-successor-id">{successor.actorId}</span>
                      {successor.role && (
                        <Badge variant="outline">{formatRoleLabel(successor.role, t)}</Badge>
                      )}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {step === 'pin' && selectedSuccessor && (
              <PinStep
                personnelId={selectedSuccessor.actorId}
                pin={pin}
                onPinChange={setPin}
                onEnter={handleAdvance}
                onChangeIdentity={handleChangeSuccessor}
                maxAttempts={MAX_ATTEMPTS}
              />
            )}

            {step === 'credential' && selectedSuccessor && (
              <CredentialStep
                personnelId={selectedSuccessor.actorId}
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
          </div>
        </div>

        {step === 'pickSuccessor' && (
          <div className="sr-button-row" style={{ marginTop: '1.5rem' }}>
            <Button type="button" variant="ghost" onClick={() => navigate('/dashboard')}>
              {t('common.cancel')}
            </Button>
            {/* Disabled here: picking a successor row above advances the
             * wizard directly (no separate confirm step) — see this file's
             * top doc comment. Kept visible so the footer still matches
             * decision #1's "Cancel + Hand off only" shape. */}
            <Button type="button" disabled>
              {t('handover.confirmHandover')}
            </Button>
          </div>
        )}
      </main>
    </div>
  );
}
