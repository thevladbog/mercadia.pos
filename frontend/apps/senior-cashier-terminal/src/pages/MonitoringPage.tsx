import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button } from '@mercadia/ui';
import { useListStoreTerminals } from '@mercadia/api-clients-store-edge';

import { useAuth } from '@/auth/AuthProvider.js';
import { getStoreId } from '@/api-client-config.js';
import { selectSuccessData } from '@/lib/cash-utils.js';
import { TerminalHeader } from '@/components/TerminalHeader.js';

const FILTERS = ['all', 'ready', 'busy', 'error', 'offline'] as const;
type Filter = (typeof FILTERS)[number];

const FILTER_I18N_KEYS: Record<Filter, string> = {
  all: 'monitoring.filterAll',
  ready: 'monitoring.filterActive',
  busy: 'monitoring.filterAttention',
  error: 'monitoring.filterBlocked',
  offline: 'monitoring.filterOffline',
};

const STATUS_COLORS: Record<string, string> = {
  ready: 'var(--ui-success)',
  busy: 'var(--ui-warning)',
  error: 'var(--ui-danger)',
  offline: 'var(--ui-text-muted)',
};

export function MonitoringPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { logout } = useAuth();
  const storeId = useMemo(() => getStoreId(), []);
  const [filter, setFilter] = useState<Filter>('all');

  const { data: terminalsResp, isFetching } = useListStoreTerminals(storeId);

  const terminalsData = selectSuccessData<{
    items: { id: string; kind: string; status: string; lastSeenAt: string }[];
  }>(terminalsResp);

  const filtered = useMemo(() => {
    const items = terminalsData?.items ?? [];
    if (filter === 'all') return items;
    return items.filter((item) => item.status === filter);
  }, [terminalsData, filter]);

  const statusColor = (status: string) => STATUS_COLORS[status] ?? 'var(--ui-text)';

  return (
    <div className="sr-terminal-shell">
      <TerminalHeader title={t('monitoring.title')} onLogout={logout} />

      <main className="sr-terminal-main">
        <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem', flexWrap: 'wrap' }}>
          {FILTERS.map((f) => (
            <Button
              key={f}
              variant={filter === f ? 'primary' : 'secondary'}
              size="sm"
              onClick={() => setFilter(f)}
            >
              {t(FILTER_I18N_KEYS[f])}
            </Button>
          ))}
        </div>

        <div className="sr-panel">
          {isFetching ? (
            <p className="muted">{t('common.loading')}</p>
          ) : !terminalsResp ? (
            <p className="muted">{t('common.loading')}</p>
          ) : filtered.length === 0 ? (
            <p className="muted">{t('monitoring.noTerminals')}</p>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {filtered.map((terminal) => (
                <div
                  key={terminal.id}
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    padding: '0.75rem',
                    border: '1px solid var(--ui-border)',
                    borderRadius: 'var(--ui-radius-md)',
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                    <span
                      style={{
                        width: 10,
                        height: 10,
                        borderRadius: '50%',
                        background: statusColor(terminal.status),
                        display: 'inline-block',
                      }}
                    />
                    <div>
                      <div style={{ fontWeight: 500 }}>{terminal.id}</div>
                      <div className="muted" style={{ fontSize: '0.85rem' }}>
                        {terminal.kind}
                      </div>
                    </div>
                  </div>
                  <div style={{ textAlign: 'right', fontSize: '0.85rem' }}>
                    <div>{terminal.status}</div>
                    {terminal.lastSeenAt && (
                      <div className="muted">
                        {new Date(terminal.lastSeenAt).toLocaleTimeString()}
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <Button
          variant="ghost"
          onClick={() => navigate('/dashboard')}
          style={{ marginTop: '1rem' }}
        >
          {t('common.back')}
        </Button>
      </main>
    </div>
  );
}
