import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { ActionCard } from './ActionCard.js';
import { TagIcon, TruckIcon } from './icons.js';

/** "ОПЕРАЦИИ С СЕЙФОМ" secondary row (plan 021). Only 2 cards — the design's
 * 3rd slot was safe-recount, now promoted into `PrimaryActionsRow`; this
 * deliberately does not backfill a 3rd, unrelated card just to fill the row. */
export function SafeOpsRow() {
  const { t } = useTranslation();
  const navigate = useNavigate();

  return (
    <div className="sr-dashboard-action-grid sr-dashboard-action-grid--secondary">
      <ActionCard
        icon={<TruckIcon />}
        title={t('dashboard.bankCollection')}
        subtitle={t('dashboard.cards.bankCollectionSubtitle')}
        onClick={() => navigate('/cash/bank-collection')}
      />
      <ActionCard
        icon={<TagIcon />}
        title={t('dashboard.expense')}
        subtitle={t('dashboard.cards.expenseSubtitle')}
        onClick={() => navigate('/cash/expense')}
      />
    </div>
  );
}
