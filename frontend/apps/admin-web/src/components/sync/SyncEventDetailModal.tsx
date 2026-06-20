import type { ListStoreSyncEvents200ItemsItem } from '@mercadia/api-clients-central';
import { useTranslation } from 'react-i18next';

import { DetailModal } from '@/components/eod/DetailModal.js';
import { formatTimestamp } from '@/pages/reporting-utils.js';

type SyncEventDetailModalProps = {
  event: ListStoreSyncEvents200ItemsItem;
  onClose: () => void;
};

function formatPayload(payload: ListStoreSyncEvents200ItemsItem['payload']): string {
  try {
    return JSON.stringify(payload, null, 2);
  } catch {
    return String(payload);
  }
}

export function SyncEventDetailModal({ event, onClose }: SyncEventDetailModalProps) {
  const { t } = useTranslation();

  return (
    <DetailModal title={t('sync.eventDetail.title')} onClose={onClose}>
      <dl className="kpi-grid">
        <div>
          <dt>{t('sync.eventDetail.eventId')}</dt>
          <dd>{event.eventId}</dd>
        </div>
        <div>
          <dt>{t('sync.eventDetail.sourceEventId')}</dt>
          <dd>{event.sourceEventId}</dd>
        </div>
        <div>
          <dt>{t('sync.eventDetail.eventType')}</dt>
          <dd>{event.eventType}</dd>
        </div>
        <div>
          <dt>{t('sync.eventDetail.occurredAt')}</dt>
          <dd>{formatTimestamp(event.occurredAt)}</dd>
        </div>
        <div>
          <dt>{t('sync.eventDetail.receivedAt')}</dt>
          <dd>{formatTimestamp(event.receivedAt)}</dd>
        </div>
      </dl>
      <h4>{t('sync.eventDetail.payload')}</h4>
      <pre className="sync-event-payload">{formatPayload(event.payload)}</pre>
    </DetailModal>
  );
}
