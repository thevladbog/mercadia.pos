import {
  getGetLayoutTemplateQueryKey,
  getListLayoutTemplatesQueryKey,
  useGetLayoutTemplate,
  useListColorSchemes,
  useListStoreCatalogProducts,
  useListStores,
  useUpdateLayoutTemplate,
  type GetLayoutTemplate200Template,
  type UpdateLayoutTemplateBody,
} from '@mercadia/api-clients-central';
import { Button } from '@mercadia/ui';
import { useQueryClient } from '@tanstack/react-query';
import { useMemo, useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from 'react-router-dom';

import { LayoutGridEditor } from '@/components/branding/LayoutGridEditor.js';
import { LayoutTemplatePreview } from '@/components/branding/LayoutTemplatePreview.js';
import { getApiErrorMessage } from '@/auth/api-errors.js';
import { SelectField, TextField } from '@/components/FormControls.js';
import {
  ACCENT_PRESET_OPTIONS,
  accentPresetLabel,
  BrandingBackLink,
} from '@/pages/branding-shared.js';
import { gridFromApi, gridToApi, validateGridForPublish } from '@/pages/layout-template-utils.js';

const KIND_OPTIONS = ['sale', 'return', 'sco'] as const;

type EditLayoutTemplateFormProps = {
  template: GetLayoutTemplate200Template;
  templateId: string;
};

function EditLayoutTemplateForm({ template, templateId }: EditLayoutTemplateFormProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const storesQuery = useListStores();
  const schemesQuery = useListColorSchemes();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const schemes = schemesQuery.data?.status === 200 ? schemesQuery.data.data.schemes : [];

  const [name, setName] = useState(template.name);
  const [kind, setKind] = useState(template.kind);
  const [accentPreset, setAccentPreset] = useState(template.accentPreset ?? '');
  const [accentColor, setAccentColor] = useState(template.accentColor ?? '');
  const [colorSchemeId, setColorSchemeId] = useState(template.colorSchemeId ?? '');
  const [storeId, setStoreId] = useState(template.storeId ?? '');
  const [terminalType, setTerminalType] = useState(template.terminalType ?? '');
  const [status, setStatus] = useState(template.status);
  const [grid, setGrid] = useState(() => gridFromApi(template.grid));
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const productsQuery = useListStoreCatalogProducts(storeId, {
    query: { enabled: storeId.length > 0 },
  });
  const catalogReady = storeId.length === 0 || productsQuery.data?.status === 200;
  const knownProductIds = useMemo(() => {
    const products = productsQuery.data?.status === 200 ? productsQuery.data.data.products : [];
    return new Set(products.map((product) => product.id));
  }, [productsQuery.data]);

  const mutation = useUpdateLayoutTemplate({
    mutation: {
      onSuccess: async (response) => {
        if (response.status !== 200) {
          setErrorMessage(t('common.unexpectedError'));
          return;
        }
        await queryClient.invalidateQueries({ queryKey: getListLayoutTemplatesQueryKey() });
        await queryClient.invalidateQueries({ queryKey: getGetLayoutTemplateQueryKey(templateId) });
        void navigate('/central/layout-templates');
      },
      onError: (error) => {
        setErrorMessage(getApiErrorMessage(error));
      },
    },
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage(null);

    if (status === 'published') {
      const validation = validateGridForPublish(grid, storeId, { catalogReady, knownProductIds });
      if (!validation.ok) {
        if (validation.reason === 'publishRequiresStore') {
          setErrorMessage(t('layoutTemplates.publishRequiresStore'));
        } else if (validation.reason === 'catalogNotReady') {
          setErrorMessage(t('layoutTemplates.catalogNotReady'));
        } else {
          setErrorMessage(
            t('layoutTemplates.invalidProducts', {
              productIds: validation.productIds?.join(', ') ?? '',
            }),
          );
        }
        return;
      }
    }

    const payload: UpdateLayoutTemplateBody = {
      name,
      kind,
      status,
      grid: gridToApi(grid),
      accentPreset,
      accentColor,
      colorSchemeId,
      storeId,
      terminalType,
    };

    mutation.mutate({ templateId, data: payload });
  }

  return (
    <div className="stack">
      <form className="stack" onSubmit={handleSubmit}>
        <p className="readonly-field">{template.id}</p>
        <TextField
          required
          label={t('layoutTemplates.name')}
          value={name}
          onValueChange={setName}
        />
        <SelectField label={t('layoutTemplates.kind')} value={kind} onValueChange={setKind}>
          {KIND_OPTIONS.map((option) => (
            <option key={option} value={option}>
              {t(`layoutTemplates.kind.${option}`)}
            </option>
          ))}
        </SelectField>
        <SelectField
          label={t('layoutTemplates.accentPreset')}
          value={accentPreset}
          onValueChange={setAccentPreset}
        >
          <option value="">{t('layoutTemplates.accentPresetDefault')}</option>
          {ACCENT_PRESET_OPTIONS.map((preset) => (
            <option key={preset} value={preset}>
              {accentPresetLabel(t, preset)}
            </option>
          ))}
        </SelectField>
        <TextField
          label={t('layoutTemplates.accentColor')}
          value={accentColor}
          onValueChange={setAccentColor}
        />
        <SelectField
          label={t('layoutTemplates.colorSchemeId')}
          value={colorSchemeId}
          onValueChange={setColorSchemeId}
        >
          <option value="">{t('layoutTemplates.noColorScheme')}</option>
          {schemes.map((scheme) => (
            <option key={scheme.id} value={scheme.id}>
              {scheme.name}
            </option>
          ))}
        </SelectField>
        <SelectField
          label={t('layoutTemplates.storeId')}
          value={storeId}
          onValueChange={setStoreId}
        >
          <option value="">{t('layoutTemplates.anyStore')}</option>
          {stores.map((store) => (
            <option key={store.id} value={store.id}>
              {store.name}
            </option>
          ))}
        </SelectField>
        <TextField
          label={t('layoutTemplates.terminalType')}
          value={terminalType}
          onValueChange={setTerminalType}
        />
        <SelectField label={t('layoutTemplates.status')} value={status} onValueChange={setStatus}>
          <option value="draft">{t('branding.status.draft')}</option>
          <option value="published">{t('branding.status.published')}</option>
        </SelectField>
        <LayoutGridEditor
          catalogReady={storeId ? catalogReady : undefined}
          grid={grid}
          knownProductIds={storeId ? knownProductIds : undefined}
          onChange={setGrid}
        />
        {errorMessage ? <p className="error">{errorMessage}</p> : null}
        <div className="form-actions">
          <Button disabled={mutation.isPending} type="submit">
            {mutation.isPending ? t('layoutTemplates.saving') : t('common.save')}
          </Button>
        </div>
      </form>
      <LayoutTemplatePreview
        accentColor={accentColor}
        accentPreset={accentPreset}
        grid={grid}
        kind={kind}
        resolvedAccentColor={template.resolvedAccentColor}
      />
    </div>
  );
}

export function EditLayoutTemplatePage() {
  const { t } = useTranslation();
  const { templateId = '' } = useParams();
  const templateQuery = useGetLayoutTemplate(templateId, {
    query: { enabled: templateId.length > 0 },
  });
  const template = templateQuery.data?.status === 200 ? templateQuery.data.data.template : null;
  const loadError = templateQuery.error != null ? getApiErrorMessage(templateQuery.error) : null;

  return (
    <section className="stack users-page">
      <BrandingBackLink label={t('layoutTemplates.backToList')} to="/central/layout-templates" />
      <div className="panel">
        <h2>{t('layoutTemplates.editTitle')}</h2>
        {loadError ? (
          <p className="error">{loadError}</p>
        ) : templateQuery.isLoading && !template ? (
          <p className="muted">{t('layoutTemplates.loading')}</p>
        ) : template ? (
          <EditLayoutTemplateForm key={template.id} template={template} templateId={templateId} />
        ) : null}
      </div>
    </section>
  );
}
