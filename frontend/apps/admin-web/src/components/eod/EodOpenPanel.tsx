import { useState } from 'react';
import { useTranslation } from 'react-i18next';

import { Button } from '@mercadia/ui';

import { OpenOperationalDayModal } from '@/components/eod/OpenOperationalDayModal.js';

type EodOpenPanelProps = {
  storeId: string;
  canWrite: boolean;
};

export function EodOpenPanel({ storeId, canWrite }: EodOpenPanelProps) {
  const { t } = useTranslation();
  const [openModalOpen, setOpenModalOpen] = useState(false);

  if (!canWrite) {
    return null;
  }

  return (
    <>
      <div className="panel">
        <h3>{t('eod.actions.openDayTitle')}</h3>
        <Button onClick={() => setOpenModalOpen(true)} type="button">
          {t('eod.actions.openDay')}
        </Button>
      </div>

      {openModalOpen ? (
        <OpenOperationalDayModal storeId={storeId} onClose={() => setOpenModalOpen(false)} />
      ) : null}
    </>
  );
}
