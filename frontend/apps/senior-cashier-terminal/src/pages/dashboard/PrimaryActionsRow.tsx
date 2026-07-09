import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { ActionCard } from './ActionCard.js';
import { BoxIcon, DownArrowIcon, EyeIcon, UpArrowIcon } from './icons.js';

/**
 * The primary 2x2 action grid (plan 021, design screen 02). The 4th slot
 * promotes "Пересчёт сейфа" (a real, existing `/cash/safe-recount` route)
 * from the design's secondary row, replacing the return/cancel-confirm
 * card — there is no backend pending-approval state for returns (plan 021
 * "Why this matters" item 5), so building that card would mean inventing
 * one. Safe-recount stays visually neutral (no green/blue/red accent) since
 * it doesn't carry one of the other 3 cards' money-flow meanings.
 */
export function PrimaryActionsRow() {
  const { t } = useTranslation();
  const navigate = useNavigate();

  return (
    <div className="sr-dashboard-action-grid sr-dashboard-action-grid--primary">
      <ActionCard
        variant="primary"
        accent="green"
        icon={<DownArrowIcon />}
        title={t('dashboard.cards.changeFundTitle')}
        subtitle={t('dashboard.cards.changeFundSubtitle')}
        onClick={() => navigate('/cash/change-fund')}
      />
      <ActionCard
        variant="primary"
        accent="blue"
        icon={<UpArrowIcon />}
        title={t('dashboard.cards.receiveCashTitle')}
        subtitle={t('dashboard.cards.receiveCashSubtitle')}
        onClick={() => navigate('/cash/receive')}
      />
      <ActionCard
        variant="primary"
        accent="red"
        icon={<BoxIcon />}
        title={t('dashboard.cards.finalCollectionTitle')}
        subtitle={t('dashboard.cards.finalCollectionSubtitle')}
        onClick={() => navigate('/cash/final-collection')}
      />
      <ActionCard
        variant="primary"
        accent="neutral"
        icon={<EyeIcon />}
        title={t('dashboard.safeRecount')}
        subtitle={t('dashboard.cards.safeRecountSubtitle')}
        onClick={() => navigate('/cash/safe-recount')}
      />
    </div>
  );
}
