import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  getListCashBalancesQueryKey,
  getListCashRecountsQueryKey,
  resolveCashRecount,
  type CreateCashRecount202Recount,
} from '@mercadia/api-clients-store-edge';
import { AvatarChip, Button, Dialog, DialogBody, DialogContent, DialogTitle } from '@mercadia/ui';

import { createIdempotencyHeaders, formatMinor } from '@/lib/cash-utils.js';
import { deriveInitials } from '@/lib/topbar-utils.js';

export interface RecountResolutionModalProps {
  open: boolean;
  onClose: () => void;
  storeId: string;
  /** The recount created moments ago by `SafeRecountPage`'s discrepancy
   * path — its real `id` (fixing the `recountId` bug, see plan 028) and its
   * `actorId` (the first signer, who created it — `session.actorId` at
   * creation time). */
  recount: CreateCashRecount202Recount;
  /** The just-authenticated second signer's actorId, from
   * `SecondSignerAuthModal`'s `onAuthenticated`. */
  secondSignerActorId: string;
  expectedMinor: number;
  countedMinor: number;
  /** The operator-typed discrepancy comment — becomes `resolutionNote`. */
  comment: string;
}

/**
 * The `08c`-equivalent review/sign-off modal (plan 028), shown once
 * `SecondSignerAuthModal` confirms a second signer. Shows both signatures —
 * the first signer (`recount.actorId`, already "signed" by having created
 * the recount) and the second (`secondSignerActorId`, confirmed the moment
 * this modal opens) — plus the same expected/counted/diff summary
 * `MismatchDialog` already computes, and the real discrepancy comment.
 *
 * **One button only** — "Sign the recount". `CashRecount.Resolve` (like
 * `Return.ConfirmPendingApproval`/`Shift.ConfirmClose`) has no decline/
 * reject outcome at all, so — per the user's confirm-only decision — no
 * decline button is built here, matching the identical limitation the
 * other two confirm flows already shipped without.
 *
 * Identity mapping (see plan 028's ground truth for the full reasoning):
 * `ResolveCashRecountCommand.ActorID` is the second signer (persisted as
 * `CashRecount.ResolvedByID`), `ApprovedByID` is the original creator
 * (`recount.actorId`, a required same-call separation-of-duties witness,
 * never persisted) — this exactly mirrors how `ConfirmCloseShift`/
 * `ConfirmReturn` both check permission against the *confirming* actor.
 */
export function RecountResolutionModal({
  open,
  onClose,
  storeId,
  recount,
  secondSignerActorId,
  expectedMinor,
  countedMinor,
  comment,
}: RecountResolutionModalProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const diff = countedMinor - expectedMinor;

  const mutation = useMutation({
    mutationFn: async () =>
      resolveCashRecount(
        storeId,
        recount.id,
        {
          actorId: secondSignerActorId,
          approvedById: recount.actorId,
          resolutionNote: comment,
        },
        { headers: createIdempotencyHeaders() },
      ),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: getListCashRecountsQueryKey(storeId) });
      await queryClient.invalidateQueries({ queryKey: getListCashBalancesQueryKey(storeId) });
      navigate('/dashboard', { replace: true });
    },
  });

  const errorMessage =
    mutation.error instanceof Error ? mutation.error.message : t('common.unexpectedError');

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next) onClose();
      }}
    >
      <DialogContent aria-describedby={undefined}>
        <DialogTitle>{t('cash.resolution.title')}</DialogTitle>
        <DialogBody>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <AvatarChip initials={deriveInitials(recount.actorId)} size="sm" />
            <div>
              <div className="sr-field-label">{t('cash.resolution.firstSignature')}</div>
              <div style={{ fontWeight: 600 }}>{recount.actorId}</div>
              <div className="muted">
                {t('cash.resolution.signedAt', {
                  time: new Date(recount.createdAt).toLocaleString(),
                })}
              </div>
            </div>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <AvatarChip initials={deriveInitials(secondSignerActorId)} size="sm" />
            <div>
              <div className="sr-field-label">{t('cash.resolution.secondSignature')}</div>
              <div style={{ fontWeight: 600 }}>{secondSignerActorId}</div>
              <div className="muted">{t('cash.resolution.confirmedNow')}</div>
            </div>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            <div className="sr-field-row">
              <span className="sr-field-label">{t('cash.mismatchExpected')}</span>
              <span>{formatMinor(expectedMinor)} ₽</span>
            </div>
            <div className="sr-field-row">
              <span className="sr-field-label">{t('cash.mismatchCounted')}</span>
              <span>{formatMinor(countedMinor)} ₽</span>
            </div>
            <div className="sr-field-row">
              <span className="sr-field-label">{t('cash.mismatchDiff')}</span>
              <span>
                {diff > 0 ? '+' : ''}
                {formatMinor(diff)} ₽
              </span>
            </div>
          </div>

          <div>
            <div className="sr-field-label">{t('cash.resolution.commentLabel')}</div>
            <p style={{ margin: 0 }}>{comment}</p>
          </div>

          {mutation.isError && <p className="sr-field-error">{errorMessage}</p>}

          <div style={{ display: 'flex', gap: '0.5rem', marginTop: '1rem' }}>
            <Button
              type="button"
              variant="primary"
              disabled={mutation.isPending}
              onClick={() => mutation.mutate()}
            >
              {mutation.isPending ? t('common.submitting') : t('cash.resolution.signButton')}
            </Button>
          </div>
        </DialogBody>
      </DialogContent>
    </Dialog>
  );
}
