import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';

import { useListCashBalances, useListOpenStoreShifts, useListStoreTerminals } from '@mercadia/api-clients-store-edge';
import { useListStores } from '@mercadia/api-clients-central';
import { SeniorCashierDashboard } from '@/components/senior-cashier/SeniorCashierDashboard.js';
import { StorePicker } from '@/components/StorePicker.js';
import { readStoreFromSearchParams } from '@/pages/store-routes.js';

export function SeniorCashierDashboardPage() {
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();
  const initialStoreId = readStoreFromSearchParams(searchParams);
  const [selectedStoreId, setSelectedStoreId] = useState<string | null>(initialStoreId);

  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const activeStoreId = selectedStoreId ?? stores[0]?.id ?? '';

  const balancesQuery = useListCashBalances(activeStoreId, { query: { enabled: activeStoreId.length > 0 } });
  const shiftsQuery = useListOpenStoreShifts(activeStoreId, { query: { enabled: activeStoreId.length > 0 } });
  const terminalsQuery = useListStoreTerminals(activeStoreId, undefined, { query: { enabled: activeStoreId.length > 0 } });

  const balances = balancesQuery.data?.status === 200 ? balancesQuery.data.data.balances : null;
  const shifts = shiftsQuery.data?.status === 200 ? shiftsQuery.data.data.shifts : null;
  const terminalsPage = terminalsQuery.data?.status === 200 ? terminalsQuery.data.data : null;

  if (!activeStoreId) {
    return (
      <div className="panel">
        <h1>{t('seniorCashier.dashboard')}</h1>
        <p className="muted">{t('common.noStores')}</p>
      </div>
    );
  }

  return (
    <div className="panel">
      <h1>{t('seniorCashier.dashboard')}</h1>
      <StorePicker stores={stores} value={activeStoreId} onChange={setSelectedStoreId} />
      <SeniorCashierDashboard
        storeId={activeStoreId}
        balances={balances}
        shifts={shifts}
        terminals={terminalsPage?.items ?? null}
      />
    </div>
  );
}
