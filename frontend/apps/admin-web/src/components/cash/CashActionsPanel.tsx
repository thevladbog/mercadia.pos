import type { ListCashBalances200BalancesItem } from '@mercadia/api-clients-store-edge';
import { Button } from '@mercadia/ui';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

import { BankCollectionModal } from '@/components/cash/BankCollectionModal.js';
import { BusinessExpenseModal } from '@/components/cash/BusinessExpenseModal.js';
import {
  CreateCashMovementModal,
  type CashMovementVariant,
} from '@/components/cash/CreateCashMovementModal.js';
import { CreateRecountModal } from '@/components/cash/CreateRecountModal.js';

type CashAction =
  | 'change_fund'
  | 'drawer_to_safe'
  | 'bank_collection'
  | 'business_expense'
  | 'recount';

type CashActionsPanelProps = {
  storeId: string;
  balances: ListCashBalances200BalancesItem[];
  canWrite: boolean;
};

export function CashActionsPanel({ storeId, balances, canWrite }: CashActionsPanelProps) {
  const { t } = useTranslation();
  const [activeAction, setActiveAction] = useState<CashAction | null>(null);

  if (!canWrite) {
    return null;
  }

  const hasContainers = balances.length > 0;

  function openMovement(variant: CashMovementVariant) {
    setActiveAction(variant);
  }

  function closeModal() {
    setActiveAction(null);
  }

  return (
    <>
      <div className="panel">
        <h3>{t('safe.actions.title')}</h3>
        {!hasContainers ? (
          <p className="muted">{t('safe.actions.noContainers')}</p>
        ) : (
          <div className="cash-actions">
            <Button onClick={() => openMovement('change_fund')} type="button">
              {t('safe.actions.issueChangeFund')}
            </Button>
            <Button
              variant="secondary"
              onClick={() => openMovement('drawer_to_safe')}
              type="button"
            >
              {t('safe.actions.receiveFromCashier')}
            </Button>
            <Button
              variant="secondary"
              onClick={() => setActiveAction('bank_collection')}
              type="button"
            >
              {t('safe.actions.bankCollection')}
            </Button>
            <Button
              variant="secondary"
              onClick={() => setActiveAction('business_expense')}
              type="button"
            >
              {t('safe.actions.businessExpense')}
            </Button>
            <Button variant="secondary" onClick={() => setActiveAction('recount')} type="button">
              {t('safe.actions.recountSafe')}
            </Button>
          </div>
        )}
      </div>

      {activeAction === 'change_fund' || activeAction === 'drawer_to_safe' ? (
        <CreateCashMovementModal
          key={activeAction}
          balances={balances}
          storeId={storeId}
          variant={activeAction}
          onClose={closeModal}
        />
      ) : null}
      {activeAction === 'bank_collection' ? (
        <BankCollectionModal
          key="bank_collection"
          balances={balances}
          storeId={storeId}
          onClose={closeModal}
        />
      ) : null}
      {activeAction === 'business_expense' ? (
        <BusinessExpenseModal
          key="business_expense"
          balances={balances}
          storeId={storeId}
          onClose={closeModal}
        />
      ) : null}
      {activeAction === 'recount' ? (
        <CreateRecountModal
          key="recount"
          balances={balances}
          storeId={storeId}
          onClose={closeModal}
        />
      ) : null}
    </>
  );
}
