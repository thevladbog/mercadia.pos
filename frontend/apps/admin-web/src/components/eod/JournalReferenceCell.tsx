import type { ListOperationJournal200ItemsItem } from '@mercadia/api-clients-store-edge';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import {
  isJournalReferenceActionable,
  resolveJournalReferenceAction,
} from '@/pages/eod-journal-utils.js';

type JournalReferenceCellProps = {
  entry: ListOperationJournal200ItemsItem;
  storeId: string;
  onOpenReceipt: (receiptId: string) => void;
  onOpenReturn: (returnId: string) => void;
  onOpenShift: (shiftId: string) => void;
};

export function JournalReferenceCell({
  entry,
  storeId,
  onOpenReceipt,
  onOpenReturn,
  onOpenShift,
}: JournalReferenceCellProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const action = resolveJournalReferenceAction(entry.operationType, entry.referenceId, storeId);

  function runAction() {
    switch (action.kind) {
      case 'navigate':
        void navigate(action.href);
        return;
      case 'receiptModal':
        onOpenReceipt(action.receiptId);
        return;
      case 'returnModal':
        onOpenReturn(action.returnId);
        return;
      case 'shiftModal':
        onOpenShift(action.shiftId);
        return;
      case 'none':
        return;
    }
  }

  if (!entry.referenceId) {
    return <span>{t('common.emDash')}</span>;
  }

  if (!isJournalReferenceActionable(action)) {
    return <span>{entry.referenceId}</span>;
  }

  return (
    <button
      aria-label={t('eod.journalReference.link', { reference: entry.referenceId })}
      className="link-button"
      type="button"
      onClick={runAction}
    >
      {entry.referenceId}
    </button>
  );
}
