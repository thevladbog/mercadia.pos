import type { ListStoreMonitoringTerminals200ItemsItem } from '@mercadia/api-clients-store-edge';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { terminalMonitoringHref } from './monitoring-routes.js';
import { terminalStatusClass, terminalStatusLabel } from './monitoring-utils.js';
import { formatMinorAmount, formatTimestamp } from './reporting-utils.js';

type TerminalCardGridProps = {
  storeId: string;
  terminals: ListStoreMonitoringTerminals200ItemsItem[];
};

export function TerminalCardGrid({ storeId, terminals }: TerminalCardGridProps) {
  const { t } = useTranslation();

  return (
    <div className="terminal-grid">
      {terminals.map((terminal) => (
        <article className="terminal-card panel" key={terminal.id}>
          <div className="terminal-card-header">
            <Link to={terminalMonitoringHref(storeId, terminal.id)}>{terminal.id}</Link>
            <span className={terminalStatusClass(terminal)}>{terminalStatusLabel(terminal)}</span>
          </div>
          <dl className="terminal-card-meta">
            <div>
              <dt>{t('monitoring.kind')}</dt>
              <dd>{terminal.kind}</dd>
            </div>
            <div>
              <dt>{t('monitoring.cashier')}</dt>
              <dd>{terminal.cashierId ?? t('common.emDash')}</dd>
            </div>
            <div>
              <dt>{t('monitoring.receipts')}</dt>
              <dd>{terminal.receiptCount}</dd>
            </div>
            <div>
              <dt>{t('monitoring.revenue')}</dt>
              <dd>{formatMinorAmount(terminal.revenueMinor)}</dd>
            </div>
            <div>
              <dt>{t('monitoring.drawer')}</dt>
              <dd>{formatMinorAmount(terminal.drawerBalanceMinor)}</dd>
            </div>
            <div>
              <dt>{t('monitoring.attention')}</dt>
              <dd>{terminal.attentionNeeded ? t('common.yes') : t('common.no')}</dd>
            </div>
            <div>
              <dt>{t('monitoring.lastSeen')}</dt>
              <dd>{formatTimestamp(terminal.lastSeenAt)}</dd>
            </div>
          </dl>
        </article>
      ))}
    </div>
  );
}
