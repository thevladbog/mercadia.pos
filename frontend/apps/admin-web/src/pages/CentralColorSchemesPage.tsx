import { Button } from '@mercadia/ui';
import { useListColorSchemes } from '@mercadia/api-clients-central';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { AccentSwatch, statusLabel } from '@/pages/branding-shared.js';
import { formatTimestamp } from '@/pages/reporting-utils.js';

export function CentralColorSchemesPage() {
  const { t } = useTranslation();
  const schemesQuery = useListColorSchemes();
  const schemes = schemesQuery.data?.status === 200 ? schemesQuery.data.data.schemes : null;
  const errorMessage = schemesQuery.error != null ? getApiErrorMessage(schemesQuery.error) : null;

  return (
    <section className="stack users-page">
      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>{t('colorSchemes.title')}</h2>
            <p className="muted">{t('colorSchemes.subtitle')}</p>
          </div>
          <div className="header-actions-inline">
            <Button
              disabled={schemesQuery.isFetching}
              onClick={() => void schemesQuery.refetch()}
              type="button"
              variant="secondary"
            >
              {schemesQuery.isFetching ? t('common.refreshing') : t('common.refresh')}
            </Button>
            <Link className="button-link" to="/central/color-schemes/new">
              {t('colorSchemes.createScheme')}
            </Link>
          </div>
        </div>

        {errorMessage ? (
          <p className="error">{errorMessage}</p>
        ) : schemesQuery.isLoading && !schemes ? (
          <p className="muted">{t('colorSchemes.loading')}</p>
        ) : schemes && schemes.length > 0 ? (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>{t('colorSchemes.name')}</th>
                  <th>{t('colorSchemes.accent')}</th>
                  <th>{t('colorSchemes.status')}</th>
                  <th>{t('colorSchemes.updated')}</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {schemes.map((scheme) => (
                  <tr key={scheme.id}>
                    <td>{scheme.name}</td>
                    <td>
                      <AccentSwatch color={scheme.resolvedAccentColor} />{' '}
                      {scheme.resolvedAccentColor}
                    </td>
                    <td>{statusLabel(t, scheme.status)}</td>
                    <td>{formatTimestamp(scheme.updatedAt)}</td>
                    <td>
                      <Link to={`/central/color-schemes/${scheme.id}`}>{t('common.edit')}</Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="muted">{t('colorSchemes.empty')}</p>
        )}
      </div>
    </section>
  );
}
