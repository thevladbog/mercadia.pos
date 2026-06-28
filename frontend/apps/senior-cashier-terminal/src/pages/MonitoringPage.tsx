import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button } from '@mercadia/ui';
import { useListStoreTerminals } from '@mercadia/api-clients-store-edge';

import { useAuth } from '@/auth/AuthProvider.js';
import { getStoreId } from '@/api-client-config.js';
import { TerminalHeader } from '@/components/TerminalHeader.js';

const FILTERS = ['all', 'active', 'attention', 'blocked', 'offline'] as const;
type Filter = (typeof FILTERS)[number];

const FILTER_I18N_KEYS: Record<Filter, string> = {
  all: 'monitoring.filterAll',
  active: 'monitoring.filterActive',
  attention: 'monitoring.filterAttention',
  blocked: 'monitoring.filterBlocked',
  offline: 'monitoring.filterOffline',
};

export function MonitoringPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { logout } = useAuth();
  const storeId = useMemo(() => getStoreId(), []);
  const [filter, setFilter] = useState<Filter>('all');

  const { data: terminalsResp } = useListStoreTerminals(storeId);

  const filtered = useMemo(() => {
    const items = (terminalsResp?.data as any)?.items ?? [];
    if (filter === 'all') return items;
    return items.filter((item: any) => item.status === filter);
  }, [terminalsResp, filter]);

  const statusColor = (status: string) => {
    switch (status) {
      case 'ready': return 'var(--ui-success)';
      case 'busy': return 'var(--ui-warning)';
      case 'error': return 'var(--ui-danger)';
      case 'offline': return 'var(--ui-text-muted)';
      default: return 'var(--ui-text)';
    }
  };

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
          {filtered.length === 0 ? (
            <p className="muted">{t('monitoring.noTerminals')}</p>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {filtered.map((terminal: any) => (
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
                      <div className="muted">{new Date(terminal.lastSeenAt).toLocaleTimeString()}</div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <Button variant="ghost" onClick={() => navigate('/dashboard')} style={{ marginTop: '1rem' }}>
          {t('common.back')}
        </Button>
      </main>
    </div>
  );
}
