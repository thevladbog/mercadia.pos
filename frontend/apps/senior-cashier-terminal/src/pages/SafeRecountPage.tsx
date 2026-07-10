import { useCallback, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button, Field, Label, Textarea } from '@mercadia/ui';
import {
  useListCashBalances,
  createCashRecount,
  getListCashBalancesQueryKey,
  getListCashRecountsQueryKey,
  type CreateCashRecount202Recount,
} from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { useAuth } from '@/auth/AuthProvider.js';
import { getStoreId } from '@/api-client-config.js';
import {
  computeDenominationTotal,
  createIdempotencyHeaders,
  formatMinor,
  selectSuccessData,
} from '@/lib/cash-utils.js';
import { useTopBarActions } from '@/lib/use-topbar-actions.js';
import { MismatchDialog } from '@/components/MismatchDialog.js';
import { SecondSignerAuthModal } from '@/components/SecondSignerAuthModal.js';
import { TopBar } from '@/components/TopBar.js';

import { DenominationGrid } from './cash-operations/DenominationGrid.js';
import { OperationChecklist } from './cash-operations/OperationChecklist.js';
import { RecountResolutionModal } from './cash-operations/RecountResolutionModal.js';
import { findSafeBalance, type CashBalanceForLookup } from './cash-operations/cash-operations-data.js';
import './cash-operations/CashOperations.css';

// Stable empty-array fallback (module scope) — see `FinalCollectionPage.tsx`'s
// identical `EMPTY_BALANCES` convention (plan 022/021).
const EMPTY_BALANCES: CashBalanceForLookup[] = [];

/**
 * Redesigned Safe Recount page (plan 028, Phase 6; design screens
 * `08a`/`08b`/`08c`). Matches the shared cash-operation visual language
 * (`DenominationGrid variant="recount"`, `OperationChecklist`) already used
 * by the 5 pages plan 022 redesigned. `BeforeAfterPanel` deliberately does
 * NOT apply here — unlike those 5 pages, a safe recount moves no cash
 * between containers; it only corrects the recorded balance against a
 * physical count. `MismatchDialog`'s existing expected/counted/diff framing
 * is the correct summary.
 *
 * Fixes two real pre-existing bugs (plan 028's ground truth):
 * - `recountId` used to be hardcoded to the placeholder string `'pending'`
 *   instead of the real ID from the create response, so `resolveCashRecount`
 *   was never actually being called with a real ID. Fixed: the real
 *   `CreateCashRecount202Recount` returned by `createCashRecount` is kept in
 *   state and its `.id` is what gets passed to `resolveCashRecount` (now via
 *   `RecountResolutionModal`).
 * - `resolutionNote` used to be a hardcoded placeholder
 *   (`'confirmed'`/`'discrepancy_recorded'`), never a real, operator-typed
 *   explanation. Fixed: a real comment textarea is now part of the
 *   discrepancy-resolution flow, and its value is what's submitted as
 *   `resolutionNote`.
 *
 * `actorId` is derived from `useAuth()`'s `session.actorId` (no more
 * free-text actor/approver inputs), and the safe container lookup uses the
 * shared `findSafeBalance` helper instead of an inline `.find()` — matching
 * every other cash-op page (plan 022 item 4/5).
 *
 * Two-person-signature flow for a discrepancy (screens `08b`/`08c`): the
 * design's "first signature" is this terminal's own signed-in senior
 * cashier (the one who counted and created the recount); the "second
 * signature" is a DIFFERENT senior cashier/admin who re-authenticates via
 * `SecondSignerAuthModal` specifically to sign off. `createCashRecount`'s
 * app-layer gate (`ErrCashRecountApprovalRequired`) requires a same-call
 * `approvedById` witness whenever there's already a mismatch at creation
 * time — so for a discrepancy, the actual `createCashRecount` call is
 * deferred until the second signer is confirmed (their actorId becomes
 * `approvedById`), rather than asking the operator to pre-emptively type a
 * free-text approver before a discrepancy is even known to exist. Once
 * created, `RecountResolutionModal` performs the real `resolveCashRecount`
 * sign-off (second signer as `actorId`/`ResolvedByID`, this terminal's own
 * actor as the required same-call `approvedById` witness — see that file's
 * doc comment for the full identity-mapping reasoning).
 */
export function SafeRecountPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { onHandover, onLock } = useTopBarActions();
  const { session } = useAuth();
  const storeId = useMemo(() => getStoreId(), []);
  const actorId = session?.actorId ?? '';

  const { data: balancesResp } = useListCashBalances(storeId);
  const balances =
    selectSuccessData<{ balances: CashBalanceForLookup[] }>(balancesResp)?.balances ??
    EMPTY_BALANCES;
  const balancesLoaded = balancesResp !== undefined;

  const safeBalanceEntry = useMemo(() => findSafeBalance(balances), [balances]);
  const safeId = safeBalanceEntry?.containerId ?? 'safe-1';
  const safeBalance = safeBalanceEntry?.balanceMinor ?? 0;

  const [billValues, setBillValues] = useState<Record<number, string>>({});
  const [coinsMinor, setCoinsMinor] = useState(0);
  const [otherMinor, setOtherMinor] = useState(0);
  const [comment, setComment] = useState('');
  const [error, setError] = useState('');
  const [showMismatch, setShowMismatch] = useState(false);
  const [showSecondSignerAuth, setShowSecondSignerAuth] = useState(false);
  const [discrepancyRecount, setDiscrepancyRecount] = useState<CreateCashRecount202Recount | null>(
    null,
  );
  const [secondSignerActorId, setSecondSignerActorId] = useState<string | null>(null);

  const countedMinor = useMemo(
    () => computeDenominationTotal(billValues, coinsMinor + otherMinor),
    [billValues, coinsMinor, otherMinor],
  );

  // No-discrepancy path: `countedMinor === safeBalance`, so `createCashRecount`
  // succeeds without an `approvedById` (the domain gate only fires on a real
  // mismatch) and there is nothing to sign off — matches the backend's own
  // `resolutionStatus: not_required` for a balanced recount.
  const createMutation = useMutation({
    mutationFn: async () =>
      createCashRecount(
        storeId,
        {
          containerType: 'safe',
          containerId: safeId,
          countedMinor,
          actorId,
        },
        { headers: createIdempotencyHeaders() },
      ),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: getListCashRecountsQueryKey(storeId) });
      await queryClient.invalidateQueries({ queryKey: getListCashBalancesQueryKey(storeId) });
      navigate('/dashboard', { replace: true });
    },
    onError: (err: Error) => setError(err?.message ?? t('common.unexpectedError')),
  });

  // Discrepancy path: fired only once a second signer has authenticated, so
  // their actorId can satisfy `createCashRecount`'s create-time
  // `approvedById` requirement for a mismatched count in a single call.
  const createDiscrepancyMutation = useMutation({
    mutationFn: async (approvedByActorId: string): Promise<CreateCashRecount202Recount> => {
      const res = await createCashRecount(
        storeId,
        {
          containerType: 'safe',
          containerId: safeId,
          countedMinor,
          actorId,
          approvedById: approvedByActorId,
        },
        { headers: createIdempotencyHeaders() },
      );
      if (res.status !== 202) {
        throw new Error(t('common.unexpectedError'));
      }
      return res.data.recount;
    },
    onSuccess: async (recount) => {
      await queryClient.invalidateQueries({ queryKey: getListCashRecountsQueryKey(storeId) });
      setDiscrepancyRecount(recount);
    },
    onError: (err: Error) => {
      setError(err?.message ?? t('common.unexpectedError'));
      setShowSecondSignerAuth(false);
      setSecondSignerActorId(null);
    },
  });

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      setError('');

      if (!balancesLoaded) {
        setError(t('common.loading'));
        return;
      }

      if (countedMinor === safeBalance) {
        createMutation.mutate();
        return;
      }

      setShowMismatch(true);
    },
    [balancesLoaded, countedMinor, safeBalance, createMutation, t],
  );

  const handleConfirmComment = useCallback(() => {
    setShowMismatch(false);
    setShowSecondSignerAuth(true);
  }, []);

  const handleSecondSignerAuthenticated = useCallback(
    (confirmedActorId: string) => {
      setSecondSignerActorId(confirmedActorId);
      setShowSecondSignerAuth(false);
      createDiscrepancyMutation.mutate(confirmedActorId);
    },
    [createDiscrepancyMutation],
  );

  return (
    <div className="sr-terminal-shell">
      <TopBar onHandover={onHandover} onLock={onLock} />

      <main className="sr-terminal-main">
        <h1 className="sr-page-title">{t('cash.safeRecountTitle')}</h1>
        <p className="sr-cash-op-intro">{t('cash.pageIntro.safeRecount')}</p>

        <form onSubmit={handleSubmit} className="sr-cash-op-form">
          <p className="muted">
            {t('dashboard.safeBalance')}: {formatMinor(safeBalance)} ₽
          </p>

          <div className="sr-cash-op-body">
            <div className="sr-cash-op-main">
              <DenominationGrid
                variant="recount"
                billValues={billValues}
                onBillValuesChange={setBillValues}
                coinsMinor={coinsMinor}
                onCoinsMinorChange={setCoinsMinor}
                otherMinor={otherMinor}
                onOtherMinorChange={setOtherMinor}
              />
            </div>

            <div className="sr-cash-op-side">
              <OperationChecklist
                steps={[
                  t('cash.checklist.safeRecount1'),
                  t('cash.checklist.safeRecount2'),
                  t('cash.checklist.safeRecount3'),
                ]}
              />
            </div>
          </div>

          {error && <p className="sr-field-error">{error}</p>}

          <div className="sr-button-row">
            <Button type="button" variant="ghost" onClick={() => navigate('/dashboard')}>
              {t('common.cancel')}
            </Button>
            <Button
              type="submit"
              disabled={
                createMutation.isPending || createDiscrepancyMutation.isPending || !balancesLoaded
              }
            >
              {createMutation.isPending || createDiscrepancyMutation.isPending
                ? t('common.submitting')
                : !balancesLoaded
                  ? t('common.loading')
                  : t('common.confirm')}
            </Button>
          </div>
        </form>

        <MismatchDialog
          expectedMinor={safeBalance}
          countedMinor={countedMinor}
          open={showMismatch}
          onClose={() => setShowMismatch(false)}
          onResolve={handleConfirmComment}
          resolveDisabled={comment.trim().length === 0}
        >
          <Field>
            <Label>{t('cash.discrepancyComment')}</Label>
            <Textarea
              rows={3}
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              placeholder={t('cash.discrepancyCommentPlaceholder')}
            />
          </Field>
        </MismatchDialog>

        <SecondSignerAuthModal
          open={showSecondSignerAuth}
          onClose={() => setShowSecondSignerAuth(false)}
          onAuthenticated={handleSecondSignerAuthenticated}
        />

        {discrepancyRecount && secondSignerActorId && (
          <RecountResolutionModal
            open
            onClose={() => {
              setDiscrepancyRecount(null);
              setSecondSignerActorId(null);
              setShowSecondSignerAuth(false);
            }}
            storeId={storeId}
            recount={discrepancyRecount}
            secondSignerActorId={secondSignerActorId}
            expectedMinor={safeBalance}
            countedMinor={countedMinor}
            comment={comment}
          />
        )}
      </main>
    </div>
  );
}
