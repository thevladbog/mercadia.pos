import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button } from '@mercadia/ui';
import { useListOperationJournal } from '@mercadia/api-clients-store-edge';

import { useAuth } from '@/auth/AuthProvider.js';
import { getStoreId } from '@/api-client-config.js';
import { TerminalHeader } from '@/components/TerminalHeader.js';

const PAGE_SIZE = 20;

export function OperationJournalPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { logout } = useAuth();
  const storeId = useMemo(() => getStoreId(), []);
  const [page, setPage] = useState(0);

  const { data: journalResp, isFetching } = useListOperationJournal(storeId, { limit: PAGE_SIZE, offset: page * PAGE_SIZE });

  const items = (journalResp?.data as any)?.items ?? [];
  const totalCount = (journalResp?.data as any)?.totalCount ?? 0;
  const totalPages = Math.ceil(totalCount / PAGE_SIZE);

  const pageItems = useMemo(() => {
    const start = page * PAGE_SIZE;
    return items.slice(start, start + PAGE_SIZE);
  }, [items, page]);

  return (
    <div className="sr-terminal-shell">
      <TerminalHeader title={t('journal.title')} onLogout={logout} />

      <main className="sr-terminal-main">
        <div className="sr-panel">
          {isFetching ? (
            <p className="muted">{t('common.loading')}</p>
          ) : pageItems.length === 0 ? (
            <p className="muted">{t('journal.noEntries')}</p>
          ) : (
            <>
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--ui-border)' }}>
                    <th style={{ textAlign: 'left', padding: '0.5rem' }}>{t('journal.timestamp')}</th>
                    <th style={{ textAlign: 'left', padding: '0.5rem' }}>{t('journal.operation')}</th>
                    <th style={{ textAlign: 'left', padding: '0.5rem' }}>{t('journal.actor')}</th>
                    <th style={{ textAlign: 'right', padding: '0.5rem' }}>{t('journal.amount')}</th>
                    <th style={{ textAlign: 'center', padding: '0.5rem' }}>{t('journal.status')}</th>
                  </tr>
                </thead>
                <tbody>
                  {pageItems.map((entry: any) => (
                    <tr key={entry.id} style={{ borderBottom: '1px solid var(--ui-border)' }}>
                      <td style={{ padding: '0.5rem', fontSize: '0.85rem' }}>
                        {new Date(entry.createdAt).toLocaleString()}
                      </td>
                      <td style={{ padding: '0.5rem' }}>{entry.operationType}</td>
                      <td style={{ padding: '0.5rem' }}>{entry.actorId}</td>
                      <td style={{ padding: '0.5rem', textAlign: 'right', fontWeight: 600 }}>
                        {'—'}
                      </td>
                      <td style={{ padding: '0.5rem', textAlign: 'center' }}>
                        {entry.operationType}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>

              {totalPages > 1 && (
                <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'center', marginTop: '1rem' }}>
                  <Button
                    variant="secondary"
                    size="sm"
                    disabled={page === 0}
                    onClick={() => setPage((p) => p - 1)}
                  >
                    {t('common.previous')}
                  </Button>
                  <span style={{ padding: '0.25rem 0.5rem' }}>
                    {page + 1} / {totalPages}
                  </span>
                  <Button
                    variant="secondary"
                    size="sm"
                    disabled={page >= totalPages - 1}
                    onClick={() => setPage((p) => p + 1)}
                  >
                    {t('common.next')}
                  </Button>
                </div>
              )}
            </>
          )}
        </div>

        <Button variant="ghost" onClick={() => navigate('/dashboard')} style={{ marginTop: '1rem' }}>
          {t('common.back')}
        </Button>
      </main>
    </div>
  );
}
