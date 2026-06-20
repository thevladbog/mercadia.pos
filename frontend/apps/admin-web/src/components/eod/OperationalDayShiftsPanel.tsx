import { useListOperationalDayShifts } from '@mercadia/api-clients-store-edge';
import { Button } from '@mercadia/ui';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { PaginationControls } from '@/components/PaginationControls.js';
import { formatMinorAmount, formatTimestamp, PAGE_SIZE } from '@/pages/reporting-utils.js';
import { STORE_POLL_INTERVAL_MS } from '@/pages/store-polling.js';

type OperationalDayShiftsPanelProps = {
  operationalDayId: string;
  onOpenShift: (shiftId: string) => void;
};

export function OperationalDayShiftsPanel({
  operationalDayId,
  onOpenShift,
}: OperationalDayShiftsPanelProps) {
  const { t } = useTranslation();
  const [offset, setOffset] = useState(0);

  const shiftsQuery = useListOperationalDayShifts(
    operationalDayId,
    { limit: PAGE_SIZE, offset },
    {
      query: {
        enabled: operationalDayId.length > 0,
        refetchInterval: STORE_POLL_INTERVAL_MS,
      },
    },
  );
  const page = shiftsQuery.data?.status === 200 ? shiftsQuery.data.data : null;
  const errorMessage = shiftsQuery.error != null ? getApiErrorMessage(shiftsQuery.error) : null;
  const emDash = t('common.emDash');

  return (
    <div className="panel">
      <h3>{t('eod.tabs.shifts')}</h3>
      {errorMessage ? <p className="error">{errorMessage}</p> : null}
      {shiftsQuery.isLoading && !page ? (
        <p className="muted">{t('eod.shifts.loading')}</p>
      ) : page && page.items.length > 0 ? (
        <>
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>{t('eod.shiftId')}</th>
                  <th>{t('monitoring.cashier')}</th>
                  <th>{t('eod.terminalId')}</th>
                  <th>{t('monitoring.status')}</th>
                  <th>{t('eod.opened')}</th>
                  <th>{t('eod.closedAt')}</th>
                  <th>{t('eod.openingCash')}</th>
                  <th>{t('eod.shiftDetail.closingCash')}</th>
                </tr>
              </thead>
              <tbody>
                {page.items.map((shift) => (
                  <tr key={shift.id}>
                    <td>
                      <Button
                        variant="link"
                        size="sm"
                        onClick={() => onOpenShift(shift.id)}
                        type="button"
                      >
                        {shift.id}
                      </Button>
                    </td>
                    <td>{shift.cashierId}</td>
                    <td>{shift.terminalId}</td>
                    <td>{shift.status}</td>
                    <td>{formatTimestamp(shift.openedAt)}</td>
                    <td>{shift.closedAt ? formatTimestamp(shift.closedAt) : emDash}</td>
                    <td>{formatMinorAmount(shift.openingCashMinor)}</td>
                    <td>{formatMinorAmount(shift.closingCashMinor)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <PaginationControls
            canGoNext={offset + PAGE_SIZE < page.totalCount}
            canGoPrev={offset > 0}
            disabled={shiftsQuery.isFetching}
            onNext={() => setOffset((value) => value + PAGE_SIZE)}
            onPrev={() => setOffset((value) => Math.max(0, value - PAGE_SIZE))}
          />
        </>
      ) : (
        <p className="muted">{t('eod.shifts.empty')}</p>
      )}
    </div>
  );
}
