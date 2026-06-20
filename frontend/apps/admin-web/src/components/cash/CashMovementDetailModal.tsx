import type { ListCashMovements200ItemsItem } from '@mercadia/api-clients-store-edge';
import { useTranslation } from 'react-i18next';

import { DetailDialog } from '@mercadia/ui';
import { formatMinorAmount, formatTimestamp } from '@/pages/reporting-utils.js';

type CashMovementDetailModalProps = {
  movement: ListCashMovements200ItemsItem;
  onClose: () => void;
};

export function CashMovementDetailModal({ movement, onClose }: CashMovementDetailModalProps) {
  const { t } = useTranslation();
  const emDash = t('common.emDash');

  return (
    <DetailDialog
      open
      title={t('safe.movementDetail.title')}
      cancelLabel={t('common.cancel')}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <dl className="kpi-grid">
        <div>
          <dt>{t('safe.movementId')}</dt>
          <dd>{movement.id}</dd>
        </div>
        <div>
          <dt>{t('common.store')}</dt>
          <dd>{movement.storeId}</dd>
        </div>
        <div>
          <dt>{t('safe.type')}</dt>
          <dd>{movement.type}</dd>
        </div>
        <div>
          <dt>{t('monitoring.status')}</dt>
          <dd>{movement.status}</dd>
        </div>
        <div>
          <dt>{t('safe.amount')}</dt>
          <dd>{formatMinorAmount(movement.amountMinor)}</dd>
        </div>
        <div>
          <dt>{t('safe.movementDetail.currency')}</dt>
          <dd>{movement.currency}</dd>
        </div>
        <div>
          <dt>{t('safe.from')}</dt>
          <dd>
            {t('safe.containerLabel', {
              containerId: movement.fromContainerId,
              containerType: movement.fromContainerType,
            })}
          </dd>
        </div>
        <div>
          <dt>{t('safe.to')}</dt>
          <dd>
            {t('safe.containerLabel', {
              containerId: movement.toContainerId,
              containerType: movement.toContainerType,
            })}
          </dd>
        </div>
        <div>
          <dt>{t('safe.actor')}</dt>
          <dd>{movement.actorId}</dd>
        </div>
        <div>
          <dt>{t('safe.movementDetail.approvedBy')}</dt>
          <dd>
            {movement.approvedById && movement.approvedById.length > 0
              ? movement.approvedById
              : emDash}
          </dd>
        </div>
        <div>
          <dt>{t('safe.movementDetail.reason')}</dt>
          <dd>{movement.reason && movement.reason.length > 0 ? movement.reason : emDash}</dd>
        </div>
        <div>
          <dt>{t('safe.posted')}</dt>
          <dd>{formatTimestamp(movement.createdAt)}</dd>
        </div>
      </dl>
    </DetailDialog>
  );
}
