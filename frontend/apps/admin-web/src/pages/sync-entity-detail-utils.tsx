import type { ReactNode } from 'react';
import { Link } from 'react-router-dom';

import { Button } from '@mercadia/ui';

import { i18n } from '@/i18n/index.js';
import { formatMinorAmount, formatTimestamp } from './reporting-utils.js';
import { syncEntityHref } from './sync-routes.js';

export function fieldLabelFromKey(key: string): string {
  return i18n.t(`sync.entityDetail.fields.${key}`, {
    defaultValue: key.replace(/([A-Z])/g, ' $1').replace(/^./, (char) => char.toUpperCase()),
  });
}

export function formatSyncEntityScalar(key: string, value: unknown): string {
  if (value == null || value === '') {
    return i18n.t('common.emDash');
  }
  if (Array.isArray(value)) {
    return value.length > 0 ? value.join(', ') : i18n.t('common.emDash');
  }
  if (typeof value === 'number' && key.toLowerCase().includes('minor')) {
    return formatMinorAmount(value);
  }
  if (
    typeof value === 'string' &&
    (key.endsWith('At') || key === 'updatedAt' || key === 'syncedAt')
  ) {
    return formatTimestamp(value);
  }
  return String(value);
}

export type SyncEntityFieldHandlers = {
  storeId: string;
  onOpenReceipt: (receiptId: string) => void;
};

export function renderSyncEntityFieldValue(
  key: string,
  value: unknown,
  handlers: SyncEntityFieldHandlers,
): ReactNode {
  if (key === 'receiptId' && typeof value === 'string' && value.length > 0) {
    return (
      <Button
        variant="link"
        size="sm"
        type="button"
        onClick={() => handlers.onOpenReceipt(value)}
        aria-label={i18n.t('monitoring.openReceiptDetails', { receiptId: value })}
      >
        {value}
      </Button>
    );
  }

  if (key === 'returnId' && typeof value === 'string' && value.length > 0) {
    return <Link to={syncEntityHref(handlers.storeId, 'returns', value)}>{value}</Link>;
  }

  if (key === 'paymentIds' && Array.isArray(value)) {
    if (value.length === 0) {
      return i18n.t('common.emDash');
    }

    return (
      <>
        {value.map((paymentId, index) => {
          const id = String(paymentId);
          return (
            <span key={id}>
              {index > 0 ? ', ' : null}
              <Link to={syncEntityHref(handlers.storeId, 'payments', id)}>{id}</Link>
            </span>
          );
        })}
      </>
    );
  }

  if (key === 'sourceEventId' && typeof value === 'string' && value.length > 0) {
    return (
      <>
        {value}
        <span className="muted"> {i18n.t('sync.entityDetail.sourceEventHint')}</span>
      </>
    );
  }

  return formatSyncEntityScalar(key, value);
}

export function fieldsFromSyncEntityRecord(
  data: Record<string, unknown>,
): { key: string; label: string; value: unknown }[] {
  return Object.entries(data).map(([key, value]) => ({
    key,
    label: fieldLabelFromKey(key),
    value,
  }));
}
