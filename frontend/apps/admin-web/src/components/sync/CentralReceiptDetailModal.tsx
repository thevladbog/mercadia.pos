import {
  useListStoreFiscalDocuments,
  useListStorePayments,
  useListStoreReturns,
} from '@mercadia/api-clients-central';
import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { DetailModal } from '@/components/eod/DetailModal.js';
import { formatMinorAmount, formatTimestamp } from '@/pages/reporting-utils.js';
import { syncEntityHref } from '@/pages/sync-routes.js';

const RELATED_ENTITY_SCAN_LIMIT = 100;

type CentralReceiptDetailModalProps = {
  storeId: string;
  receiptId: string;
  onClose: () => void;
};

export function CentralReceiptDetailModal({
  storeId,
  receiptId,
  onClose,
}: CentralReceiptDetailModalProps) {
  const { t } = useTranslation();
  const listParams = useMemo(() => ({ limit: RELATED_ENTITY_SCAN_LIMIT, offset: 0 }), []);
  const queryOptions = useMemo(
    () => ({
      query: {
        enabled: storeId.length > 0 && receiptId.length > 0,
      },
    }),
    [receiptId, storeId],
  );

  const paymentsQuery = useListStorePayments(storeId, listParams, queryOptions);
  const fiscalDocumentsQuery = useListStoreFiscalDocuments(storeId, listParams, queryOptions);
  const returnsQuery = useListStoreReturns(storeId, listParams, queryOptions);

  const payments = useMemo(() => {
    if (paymentsQuery.data?.status !== 200) {
      return [];
    }
    return paymentsQuery.data.data.items.filter((item) => item.receiptId === receiptId);
  }, [paymentsQuery.data, receiptId]);

  const fiscalDocuments = useMemo(() => {
    if (fiscalDocumentsQuery.data?.status !== 200) {
      return [];
    }
    return fiscalDocumentsQuery.data.data.items.filter((item) => item.receiptId === receiptId);
  }, [fiscalDocumentsQuery.data, receiptId]);

  const returns = useMemo(() => {
    if (returnsQuery.data?.status !== 200) {
      return [];
    }
    return returnsQuery.data.data.items.filter((item) => item.receiptId === receiptId);
  }, [returnsQuery.data, receiptId]);

  const isLoading =
    paymentsQuery.isFetching || fiscalDocumentsQuery.isFetching || returnsQuery.isFetching;
  const errorMessage =
    paymentsQuery.error != null
      ? getApiErrorMessage(paymentsQuery.error)
      : fiscalDocumentsQuery.error != null
        ? getApiErrorMessage(fiscalDocumentsQuery.error)
        : returnsQuery.error != null
          ? getApiErrorMessage(returnsQuery.error)
          : null;

  const hasRelatedEntities =
    payments.length > 0 || fiscalDocuments.length > 0 || returns.length > 0;
  const scanMayBePartial =
    (paymentsQuery.data?.status === 200 &&
      paymentsQuery.data.data.totalCount > RELATED_ENTITY_SCAN_LIMIT) ||
    (fiscalDocumentsQuery.data?.status === 200 &&
      fiscalDocumentsQuery.data.data.totalCount > RELATED_ENTITY_SCAN_LIMIT) ||
    (returnsQuery.data?.status === 200 &&
      returnsQuery.data.data.totalCount > RELATED_ENTITY_SCAN_LIMIT);

  return (
    <DetailModal title={t('sync.receiptDetail.title')} onClose={onClose}>
      <dl className="kpi-grid">
        <div>
          <dt>{t('sync.receiptDetail.receiptId')}</dt>
          <dd>{receiptId}</dd>
        </div>
        <div>
          <dt>{t('common.store')}</dt>
          <dd>{storeId}</dd>
        </div>
      </dl>

      <p className="muted">{t('sync.receiptDetail.centralHint')}</p>

      {isLoading && !hasRelatedEntities ? (
        <p className="muted">{t('common.loading')}</p>
      ) : errorMessage ? (
        <p className="error">{errorMessage}</p>
      ) : hasRelatedEntities ? (
        <div className="stack">
          {payments.length > 0 ? (
            <section>
              <h4>{t('sync.receiptDetail.payments')}</h4>
              <ul>
                {payments.map((payment) => (
                  <li key={payment.id}>
                    <Link to={syncEntityHref(storeId, 'payments', payment.id)}>{payment.id}</Link>{' '}
                    {t('common.emDash')} {payment.method}, {formatMinorAmount(payment.amountMinor)},{' '}
                    {payment.status}, {formatTimestamp(payment.capturedAt)}
                  </li>
                ))}
              </ul>
            </section>
          ) : null}

          {fiscalDocuments.length > 0 ? (
            <section>
              <h4>{t('sync.receiptDetail.fiscalDocuments')}</h4>
              <ul>
                {fiscalDocuments.map((document) => (
                  <li key={document.id}>
                    <Link to={syncEntityHref(storeId, 'fiscal-documents', document.id)}>
                      {document.id}
                    </Link>{' '}
                    {t('common.emDash')} {document.kind}, {formatMinorAmount(document.amountMinor)},{' '}
                    {formatTimestamp(document.fiscalizedAt)}
                  </li>
                ))}
              </ul>
            </section>
          ) : null}

          {returns.length > 0 ? (
            <section>
              <h4>{t('sync.receiptDetail.returns')}</h4>
              <ul>
                {returns.map((returnItem) => (
                  <li key={returnItem.id}>
                    <Link to={syncEntityHref(storeId, 'returns', returnItem.id)}>
                      {returnItem.id}
                    </Link>{' '}
                    {t('common.emDash')} {formatMinorAmount(returnItem.totalMinor)},{' '}
                    {formatTimestamp(returnItem.settledAt)}
                  </li>
                ))}
              </ul>
            </section>
          ) : null}

          {scanMayBePartial ? <p className="muted">{t('sync.receiptDetail.partialScan')}</p> : null}
        </div>
      ) : (
        <p className="muted">{t('sync.receiptDetail.noRelated')}</p>
      )}
    </DetailModal>
  );
}
