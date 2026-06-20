import { Button } from '@mercadia/ui';
import { useListLayoutTemplates } from '@mercadia/api-clients-central';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { AccentSwatch, kindLabel, statusLabel } from '@/pages/branding-shared.js';
import { formatTimestamp } from '@/pages/reporting-utils.js';

export function CentralLayoutTemplatesPage() {
  const { t } = useTranslation();
  const templatesQuery = useListLayoutTemplates();
  const templates = templatesQuery.data?.status === 200 ? templatesQuery.data.data.templates : null;
  const errorMessage =
    templatesQuery.error != null ? getApiErrorMessage(templatesQuery.error) : null;

  return (
    <section className="stack users-page">
      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>{t('layoutTemplates.title')}</h2>
            <p className="muted">{t('layoutTemplates.subtitle')}</p>
          </div>
          <div className="header-actions-inline">
            <Button
              disabled={templatesQuery.isFetching}
              onClick={() => void templatesQuery.refetch()}
              type="button"
              variant="secondary"
            >
              {templatesQuery.isFetching ? t('common.refreshing') : t('common.refresh')}
            </Button>
            <Link className="button-link" to="/central/layout-templates/new">
              {t('layoutTemplates.createTemplate')}
            </Link>
          </div>
        </div>

        {errorMessage ? (
          <p className="error">{errorMessage}</p>
        ) : templatesQuery.isLoading && !templates ? (
          <p className="muted">{t('layoutTemplates.loading')}</p>
        ) : templates && templates.length > 0 ? (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>{t('layoutTemplates.name')}</th>
                  <th>{t('layoutTemplates.kind')}</th>
                  <th>{t('layoutTemplates.accent')}</th>
                  <th>{t('layoutTemplates.status')}</th>
                  <th>{t('layoutTemplates.updated')}</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {templates.map((template) => (
                  <tr key={template.id}>
                    <td>{template.name}</td>
                    <td>{kindLabel(t, template.kind)}</td>
                    <td>
                      <AccentSwatch color={template.resolvedAccentColor} />{' '}
                      {template.resolvedAccentColor}
                    </td>
                    <td>{statusLabel(t, template.status)}</td>
                    <td>{formatTimestamp(template.updatedAt)}</td>
                    <td>
                      <Link to={`/central/layout-templates/${template.id}`}>
                        {t('common.edit')}
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="muted">{t('layoutTemplates.empty')}</p>
        )}
      </div>
    </section>
  );
}
