import type { ListCashRecounts200ItemsItem } from '@mercadia/api-clients-store-edge';
import { useTranslation } from 'react-i18next';

import { DetailDialog } from '@mercadia/ui';
import { formatMinorAmount, formatTimestamp } from '@/pages/reporting-utils.js';

type CashRecountDetailModalProps = {
  recount: ListCashRecounts200ItemsItem;
  onClose: () => void;
};

export function CashRecountDetailModal({ recount, onClose }: CashRecountDetailModalProps) {
  const { t } = useTranslation();
  const emDash = t('common.emDash');

  return (
    <DetailDialog
      open
      title={t('safe.recountDetail.title')}
      cancelLabel={t('common.cancel')}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <dl className="kpi-grid">
        <div>
          <dt>{t('safe.recountId')}</dt>
          <dd>{recount.id}</dd>
        </div>
        <div>
          <dt>{t('common.store')}</dt>
          <dd>{recount.storeId}</dd>
        </div>
        <div>
          <dt>{t('safe.container')}</dt>
          <dd>
            {t('safe.containerLabel', {
              containerId: recount.containerId,
              containerType: recount.containerType,
            })}
          </dd>
        </div>
        <div>
          <dt>{t('monitoring.status')}</dt>
          <dd>{recount.status}</dd>
        </div>
        <div>
          <dt>{t('safe.resolution')}</dt>
          <dd>{recount.resolutionStatus}</dd>
        </div>
        <div>
          <dt>{t('safe.expected')}</dt>
          <dd>{formatMinorAmount(recount.expectedMinor)}</dd>
        </div>
        <div>
          <dt>{t('safe.counted')}</dt>
          <dd>{formatMinorAmount(recount.countedMinor)}</dd>
        </div>
        <div>
          <dt>{t('safe.variance')}</dt>
          <dd>{formatMinorAmount(recount.discrepancyMinor)}</dd>
        </div>
        <div>
          <dt>{t('safe.recountDetail.currency')}</dt>
          <dd>{recount.currency}</dd>
        </div>
        <div>
          <dt>{t('safe.actor')}</dt>
          <dd>{recount.actorId}</dd>
        </div>
        <div>
          <dt>{t('safe.recountDetail.approvedBy')}</dt>
          <dd>
            {recount.approvedById && recount.approvedById.length > 0
              ? recount.approvedById
              : emDash}
          </dd>
        </div>
        <div>
          <dt>{t('safe.recountDetail.reason')}</dt>
          <dd>{recount.reason && recount.reason.length > 0 ? recount.reason : emDash}</dd>
        </div>
        <div>
          <dt>{t('safe.recountDetail.resolutionNote')}</dt>
          <dd>
            {recount.resolutionNote && recount.resolutionNote.length > 0
              ? recount.resolutionNote
              : emDash}
          </dd>
        </div>
        <div>
          <dt>{t('safe.recountDetail.resolvedBy')}</dt>
          <dd>
            {recount.resolvedById && recount.resolvedById.length > 0
              ? recount.resolvedById
              : emDash}
          </dd>
        </div>
        <div>
          <dt>{t('safe.recountDetail.resolvedAt')}</dt>
          <dd>
            {recount.resolvedAt && recount.resolvedAt.length > 0
              ? formatTimestamp(recount.resolvedAt)
              : emDash}
          </dd>
        </div>
        <div>
          <dt>{t('eod.created')}</dt>
          <dd>{formatTimestamp(recount.createdAt)}</dd>
        </div>
      </dl>
    </DetailDialog>
  );
}
