import {
  getListLayoutTemplatesQueryKey,
  useCreateLayoutTemplate,
  useListColorSchemes,
  useListStoreCatalogProducts,
  useListStores,
  type CreateLayoutTemplateBody,
} from '@mercadia/api-clients-central';
import { Button } from '@mercadia/ui';
import { useQueryClient } from '@tanstack/react-query';
import { useMemo, useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { LayoutGridEditor } from '@/components/branding/LayoutGridEditor.js';
import { LayoutTemplatePreview } from '@/components/branding/LayoutTemplatePreview.js';
import { getApiErrorMessage } from '@/auth/api-errors.js';
import { SelectField, TextField } from '@/components/FormControls.js';
import {
  ACCENT_PRESET_OPTIONS,
  accentPresetLabel,
  BrandingBackLink,
} from '@/pages/branding-shared.js';
import { defaultGrid, gridToApi, validateGridForPublish } from '@/pages/layout-template-utils.js';

const KIND_OPTIONS = ['sale', 'return', 'sco'] as const;

export function CreateLayoutTemplatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const storesQuery = useListStores();
  const schemesQuery = useListColorSchemes();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const schemes = schemesQuery.data?.status === 200 ? schemesQuery.data.data.schemes : [];

  const [templateId, setTemplateId] = useState('');
  const [name, setName] = useState('');
  const [kind, setKind] = useState<string>('sale');
  const [accentPreset, setAccentPreset] = useState('');
  const [accentColor, setAccentColor] = useState('');
  const [colorSchemeId, setColorSchemeId] = useState('');
  const [storeId, setStoreId] = useState('');
  const [terminalType, setTerminalType] = useState('');
  const [status, setStatus] = useState('draft');
  const [grid, setGrid] = useState(() => defaultGrid());
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const productsQuery = useListStoreCatalogProducts(storeId, {
    query: { enabled: storeId.length > 0 },
  });
  const catalogReady = storeId.length === 0 || productsQuery.data?.status === 200;
  const knownProductIds = useMemo(() => {
    const products = productsQuery.data?.status === 200 ? productsQuery.data.data.products : [];
    return new Set(products.map((product) => product.id));
  }, [productsQuery.data]);

  const mutation = useCreateLayoutTemplate({
    mutation: {
      onSuccess: async (response) => {
        if (response.status !== 201) {
          setErrorMessage(t('common.unexpectedError'));
          return;
        }
        await queryClient.invalidateQueries({ queryKey: getListLayoutTemplatesQueryKey() });
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

    const payload: CreateLayoutTemplateBody = {
      templateId: templateId.trim(),
      name: name.trim(),
      kind,
      status,
      grid: gridToApi(grid),
      ...(accentColor.trim()
        ? { accentColor: accentColor.trim() }
        : accentPreset
          ? { accentPreset }
          : {}),
      ...(colorSchemeId ? { colorSchemeId } : {}),
      ...(storeId ? { storeId } : {}),
      ...(terminalType.trim() ? { terminalType: terminalType.trim() } : {}),
    };

    mutation.mutate({ data: payload });
  }

  return (
    <section className="stack users-page">
      <BrandingBackLink label={t('layoutTemplates.backToList')} to="/central/layout-templates" />
      <div className="panel">
        <h2>{t('layoutTemplates.createTitle')}</h2>
        <div className="stack">
          <form className="stack" onSubmit={handleSubmit}>
            <TextField
              required
              label={t('layoutTemplates.templateId')}
              value={templateId}
              onValueChange={setTemplateId}
            />
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
            <SelectField
              label={t('layoutTemplates.status')}
              value={status}
              onValueChange={setStatus}
            >
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
                {mutation.isPending
                  ? t('layoutTemplates.creating')
                  : t('layoutTemplates.createTemplate')}
              </Button>
            </div>
          </form>
          <LayoutTemplatePreview
            accentColor={accentColor}
            accentPreset={accentPreset}
            grid={grid}
            kind={kind}
          />
        </div>
      </div>
    </section>
  );
}
