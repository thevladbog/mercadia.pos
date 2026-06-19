import type { GetOperationalDaySummary200BlockersItem } from '@mercadia/api-clients-store-edge';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

import { CloseOperationalDayModal } from '@/components/eod/CloseOperationalDayModal.js';
import { analyzeCloseReadiness } from '@/pages/eod-mutation-utils.js';

type EodActionsPanelProps = {
  storeId: string;
  operationalDayId: string;
  blockers: GetOperationalDaySummary200BlockersItem[];
  canWrite: boolean;
};

export function EodActionsPanel({
  storeId,
  operationalDayId,
  blockers,
  canWrite,
}: EodActionsPanelProps) {
  const { t } = useTranslation();
  const [closeModalOpen, setCloseModalOpen] = useState(false);

  if (!canWrite) {
    return null;
  }

  const readiness = analyzeCloseReadiness(blockers);
  const canAttemptClose = readiness.canCloseDirectly || readiness.canCloseWithOverride;
  const hint = readiness.isBlocked
    ? t('eod.actions.closeDayHintBlocked')
    : readiness.canCloseWithOverride
      ? t('eod.actions.closeDayHintOverride')
      : null;

  return (
    <>
      <div className="panel">
        <h3>{t('eod.actions.title')}</h3>
        {hint ? <p className="muted">{hint}</p> : null}
        <button disabled={!canAttemptClose} onClick={() => setCloseModalOpen(true)} type="button">
          {t('eod.actions.closeDay')}
        </button>
      </div>

      {closeModalOpen ? (
        <CloseOperationalDayModal
          operationalDayId={operationalDayId}
          storeId={storeId}
          onClose={() => setCloseModalOpen(false)}
        />
      ) : null}
    </>
  );
}
