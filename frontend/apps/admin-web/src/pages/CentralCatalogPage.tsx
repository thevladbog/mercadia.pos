import { useListStoreCatalogProducts, useListStores } from '@mercadia/api-clients-central';
import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { StorePicker } from '@/components/StorePicker.js';
import { formatMinorAmount, formatTimestamp } from './reporting-utils.js';

function matchesProductSearch(
  product: {
    id: string;
    name: string;
    barcodes: string[];
    taxCategoryId: string;
  },
  query: string,
): boolean {
  if (query.length === 0) {
    return true;
  }
  const haystack = [product.id, product.name, product.taxCategoryId, ...product.barcodes]
    .join(' ')
    .toLowerCase();
  return haystack.includes(query);
}

export function CentralCatalogPage() {
  const { t } = useTranslation();

  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const [selectedStoreId, setSelectedStoreId] = useState<string | null>(null);
  const activeStoreId = selectedStoreId ?? stores[0]?.id ?? '';
  const [searchQuery, setSearchQuery] = useState('');

  const productsQuery = useListStoreCatalogProducts(activeStoreId, {
    query: { enabled: activeStoreId.length > 0 },
  });
  const products = productsQuery.data?.status === 200 ? productsQuery.data.data.products : null;
  const normalizedSearch = searchQuery.trim().toLowerCase();
  const filteredProducts = useMemo(
    () => products?.filter((product) => matchesProductSearch(product, normalizedSearch)) ?? null,
    [products, normalizedSearch],
  );

  const isLoading =
    storesQuery.isFetching || (activeStoreId.length > 0 && productsQuery.isFetching);
  const errorMessage =
    storesQuery.error != null
      ? getApiErrorMessage(storesQuery.error)
      : productsQuery.error != null
        ? getApiErrorMessage(productsQuery.error)
        : null;

  function refetchAll() {
    void storesQuery.refetch();
    if (activeStoreId.length > 0) {
      void productsQuery.refetch();
    }
  }

  return (
    <section className="stack monitoring-page">
      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>{t('catalog.title')}</h2>
            <p className="muted">{t('catalog.subtitle')}</p>
          </div>
          <button className="secondary" disabled={isLoading} onClick={refetchAll} type="button">
            {isLoading ? t('common.refreshing') : t('common.refresh')}
          </button>
        </div>

        <StorePicker
          loading={storesQuery.isLoading}
          stores={stores}
          value={activeStoreId}
          onChange={setSelectedStoreId}
        />
      </div>

      {errorMessage ? (
        <div className="panel error-panel">
          <p className="error">{errorMessage}</p>
        </div>
      ) : null}

      {!activeStoreId ? (
        <div className="panel">
          <p className="muted">{t('catalog.selectStore')}</p>
        </div>
      ) : (
        <div className="panel">
          <label className="field terminal-search">
            <span>{t('catalog.searchHint')}</span>
            <input
              placeholder={t('catalog.searchPlaceholder')}
              type="search"
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
            />
          </label>

          {productsQuery.isLoading && !products ? (
            <p className="muted">{t('catalog.loadingProducts')}</p>
          ) : filteredProducts && filteredProducts.length > 0 ? (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>{t('catalog.productId')}</th>
                    <th>{t('catalog.name')}</th>
                    <th>{t('catalog.barcodes')}</th>
                    <th>{t('catalog.unitPrice')}</th>
                    <th>{t('catalog.taxCategory')}</th>
                    <th>{t('catalog.active')}</th>
                    <th>{t('catalog.version')}</th>
                    <th>{t('catalog.updated')}</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredProducts.map((product) => (
                    <tr key={product.id}>
                      <td>{product.id}</td>
                      <td>{product.name}</td>
                      <td>{product.barcodes.join(', ') || t('common.emDash')}</td>
                      <td>{formatMinorAmount(product.unitPriceMinor)}</td>
                      <td>{product.taxCategoryId}</td>
                      <td>{product.active ? t('common.yes') : t('common.no')}</td>
                      <td>{product.version}</td>
                      <td>{formatTimestamp(product.updatedAt)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="muted">{t('catalog.noProducts')}</p>
          )}
        </div>
      )}
    </section>
  );
}
