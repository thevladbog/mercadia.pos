import type { GetOperationalDaySummary200BlockersItem } from '@mercadia/api-clients-store-edge';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import {
  blockerActionHintKey,
  isBlockerActionable,
  resolveBlockerHintAction,
  resolveBlockerReferenceAction,
  type EodTab,
} from '@/pages/eod-blocker-utils.js';

type BlockerReferenceCellProps = {
  blocker: GetOperationalDaySummary200BlockersItem;
  storeId: string;
  onEodTab: (tab: EodTab) => void;
  onOpenReceipt: (receiptId: string) => void;
  onOpenShift: (shiftId: string) => void;
};

export function BlockerReferenceCell({
  blocker,
  storeId,
  onEodTab,
  onOpenReceipt,
  onOpenShift,
}: BlockerReferenceCellProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const action = resolveBlockerReferenceAction(blocker.code, blocker.referenceId, storeId);
  const hintKey = blockerActionHintKey(blocker.code);

  function runAction() {
    switch (action.kind) {
      case 'eodTab':
        onEodTab(action.tab);
        return;
      case 'navigate':
        void navigate(action.href);
        return;
      case 'receiptModal':
        onOpenReceipt(action.receiptId);
        return;
      case 'shiftModal':
        onOpenShift(action.shiftId);
        return;
      case 'none':
        return;
    }
  }

  if (!isBlockerActionable(action, blocker.code)) {
    return <span>{blocker.referenceId ?? t('common.emDash')}</span>;
  }

  const label = hintKey ? t(hintKey) : (blocker.referenceId ?? t('common.emDash'));

  return (
    <button
      className="link-button"
      type="button"
      onClick={runAction}
      aria-label={t('eod.blockerActions.referenceLink', {
        reference: blocker.referenceId ?? blocker.code,
      })}
    >
      {blocker.referenceId ?? label}
    </button>
  );
}

type BlockerActionCellProps = {
  blocker: GetOperationalDaySummary200BlockersItem;
  storeId: string;
  onEodTab: (tab: EodTab) => void;
  onOpenReceipt: (receiptId: string) => void;
  onOpenShift: (shiftId: string) => void;
};

export function BlockerActionCell({
  blocker,
  storeId,
  onEodTab,
  onOpenReceipt,
  onOpenShift,
}: BlockerActionCellProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const action = resolveBlockerHintAction(blocker.code, blocker.referenceId, storeId);
  const hintKey = blockerActionHintKey(blocker.code);

  if (!hintKey || !isBlockerActionable(action, blocker.code)) {
    return <span>{t('eod.blockerActions.none')}</span>;
  }

  function runAction() {
    switch (action.kind) {
      case 'eodTab':
        onEodTab(action.tab);
        return;
      case 'navigate':
        void navigate(action.href);
        return;
      case 'receiptModal':
        onOpenReceipt(action.receiptId);
        return;
      case 'shiftModal':
        onOpenShift(action.shiftId);
        return;
      case 'none':
        return;
    }
  }

  return (
    <button className="secondary" type="button" onClick={runAction}>
      {t(hintKey)}
    </button>
  );
}
