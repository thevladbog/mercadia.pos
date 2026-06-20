import { useListStores } from '@mercadia/api-clients-central';
import { Button } from '@mercadia/ui';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { useAuth } from '@/auth/useAuth.js';
import { canManageCentralUsers } from '@/auth/permissions.js';
import { formatTimestamp } from './reporting-utils.js';

export function CentralStoresPage() {
  const { t } = useTranslation();
  const { roles } = useAuth();
  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : null;
  const errorMessage = storesQuery.error != null ? getApiErrorMessage(storesQuery.error) : null;
  const canRegister = canManageCentralUsers(roles);

  return (
    <section className="stack users-page">
      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>{t('stores.title')}</h2>
            <p className="muted">{t('stores.subtitle')}</p>
          </div>
          <div className="header-actions-inline">
            <Button
              variant="secondary"
              disabled={storesQuery.isFetching}
              onClick={() => void storesQuery.refetch()}
              type="button"
            >
              {storesQuery.isFetching ? t('common.refreshing') : t('common.refresh')}
            </Button>
            {canRegister ? (
              <Link className="button-link" to="/central/stores/new">
                {t('stores.registerStore')}
              </Link>
            ) : null}
          </div>
        </div>

        {errorMessage ? (
          <p className="error">{errorMessage}</p>
        ) : storesQuery.isLoading && !stores ? (
          <p className="muted">{t('stores.loadingStores')}</p>
        ) : stores && stores.length > 0 ? (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>{t('stores.storeId')}</th>
                  <th>{t('stores.name')}</th>
                  <th>{t('stores.region')}</th>
                  <th>{t('stores.created')}</th>
                  <th>{t('stores.updated')}</th>
                </tr>
              </thead>
              <tbody>
                {stores.map((store) => (
                  <tr key={store.id}>
                    <td>{store.id}</td>
                    <td>{store.name}</td>
                    <td>{store.region || t('common.emDash')}</td>
                    <td>{formatTimestamp(store.registeredAt)}</td>
                    <td>{formatTimestamp(store.updatedAt)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="muted">{t('stores.noStores')}</p>
        )}
      </div>
    </section>
  );
}
